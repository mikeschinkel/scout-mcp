package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*DetectCurrentProjectTool)(nil)

func init() {
	mcputil.RegisterTool(&DetectCurrentProjectTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "detect_current_project",
			Description: "Detect the most recently active project by analyzing subdirectory modification times in allowed paths",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				MaxProjectsProperty,
				IgnoreGitProperty,
			},
		}),
	})
}

type DetectCurrentProjectTool struct {
	*mcputil.ToolBase
}

type ProjectInfo struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	LastModified time.Time `json:"last_modified"`
	RelativeAge  string    `json:"relative_age"`
}

var _ ToolResult = (*ProjectDetectionResult)(nil)

type ProjectDetectionResult struct {
	CurrentProject *ProjectInfo  `json:"current_project,omitempty"`
	RecentProjects []ProjectInfo `json:"recent_projects"`
	RequiresChoice bool          `json:"requires_choice"`
	ChoiceMessage  string        `json:"choice_message,omitempty"`
	Summary        string        `json:"summary"`
}

func (t *ProjectDetectionResult) ToolResult() {
}

func (t *ProjectDetectionResult) Value() string {
	jsonData, _ := json.Marshal(t)
	return string(jsonData)
}

func (t *DetectCurrentProjectTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var maxProjects int
	var ignoreGitRequirement bool
	var detectionResult ProjectDetectionResult

	logger.Info("Tool called", "tool", "detect_current_project")

	// Parse parameters
	maxProjects, err = MaxProjectsProperty.Int(req)
	if err != nil {
		goto end
	}

	ignoreGitRequirement, err = IgnoreGitProperty.Bool(req)
	if err != nil {
		goto end
	}

	logger.Info("Tool arguments parsed", "tool", "detect_current_project",
		"max_projects", maxProjects, "ignore_git_requirement", ignoreGitRequirement)

	detectionResult, err = t.detectCurrentProject(maxProjects, ignoreGitRequirement)
	if err != nil {
		goto end
	}

	logger.Info("Tool completed", "tool", "detect_current_project", "success", true)
	result = mcputil.NewToolResultJSON(detectionResult)

end:
	return result, err
}

func (t *DetectCurrentProjectTool) detectCurrentProject(maxProjects int, ignoreGitRequirement bool) (detectionResult ProjectDetectionResult, err error) {
	var allowedPaths []string
	var allProjects []ProjectInfo
	var currentProject *ProjectInfo
	var mostRecent ProjectInfo
	var secondMostRecent ProjectInfo
	var timeDiff time.Duration

	// Get allowed paths from config
	allowedPaths = t.Config().AllowedPaths()
	if len(allowedPaths) == 0 {
		err = fmt.Errorf("no allowed paths configured")
		goto end
	}

	// Scan all allowed paths for projects
	allProjects, err = t.scanAllowedPathsForProjects(allowedPaths, ignoreGitRequirement)
	if err != nil {
		goto end
	}

	// Sort by modification time (most recent first)
	sort.Slice(allProjects, func(i, j int) bool {
		return allProjects[i].LastModified.After(allProjects[j].LastModified)
	})

	// Limit to maxProjects
	if len(allProjects) > maxProjects {
		allProjects = allProjects[:maxProjects]
	}

	// Add relative age strings
	for i := range allProjects {
		allProjects[i].RelativeAge = t.formatRelativeTime(allProjects[i].LastModified)
	}

	// Detect current project logic
	if len(allProjects) == 0 {
		detectionResult = ProjectDetectionResult{
			RecentProjects: []ProjectInfo{}, // Ensure empty array, not null
			Summary:        "No projects found in allowed paths",
		}
		goto end
	}

	if len(allProjects) == 1 {
		// Only one project, that's the current one
		currentProject = &allProjects[0]
		detectionResult = ProjectDetectionResult{
			CurrentProject: currentProject,
			RecentProjects: t.excludeCurrentProject(allProjects, currentProject),
			Summary:        fmt.Sprintf("Current project: %s (only project found)", currentProject.Name),
		}
		goto end
	}

	// Check if the most recent is 24+ hours newer than the second most recent
	mostRecent = allProjects[0]
	secondMostRecent = allProjects[1]
	timeDiff = mostRecent.LastModified.Sub(secondMostRecent.LastModified)

	if timeDiff >= 24*time.Hour {
		// Clear winner - most recent is 24+ hours newer
		currentProject = &mostRecent
		detectionResult = ProjectDetectionResult{
			CurrentProject: currentProject,
			RecentProjects: t.excludeCurrentProject(allProjects, currentProject),
			Summary:        fmt.Sprintf("Current project: %s (last modified %s)", currentProject.Name, currentProject.RelativeAge),
		}
	} else {
		// Multiple recent projects - user needs to choose
		detectionResult = ProjectDetectionResult{
			RecentProjects: allProjects,
			RequiresChoice: true,
			ChoiceMessage:  fmt.Sprintf("Multiple projects modified recently (within 24 hours). Most recent %d projects:", len(allProjects)),
			Summary:        "Multiple recent projects found - user choice required",
		}
	}

end:
	return detectionResult, err
}

func (t *DetectCurrentProjectTool) scanAllowedPathsForProjects(allowedPaths []string, ignoreGitRequirement bool) (projects []ProjectInfo, err error) {
	var foundProjects map[string]bool

	foundProjects = make(map[string]bool)

	for _, allowedPath := range allowedPaths {
		var pathProjects []ProjectInfo

		pathProjects, err = t.scanSinglePathForProjects(allowedPath, ignoreGitRequirement)
		if err != nil {
			// Log error but continue with other paths
			logger.Error("Failed to scan path for projects", "path", allowedPath, "error", err)
			continue
		}

		// Add projects, but avoid duplicates
		for _, project := range pathProjects {
			if !foundProjects[project.Path] {
				projects = append(projects, project)
				foundProjects[project.Path] = true
			}
		}
	}

	return projects, nil
}

func (t *DetectCurrentProjectTool) scanSinglePathForProjects(basePath string, ignoreGitRequirement bool) (projects []ProjectInfo, err error) {
	var entries []os.DirEntry
	var recentFileTime time.Time
	var fileCount int
	var foundProjects map[string]bool

	foundProjects = make(map[string]bool)

	// First, check if the basePath itself is a project (has recent file modifications and enough files)
	recentFileTime, fileCount, err = t.findMostRecentFileTimeAndCount(basePath)
	if err == nil && fileCount >= 5 {
		// Check if basePath has .git directory or ignore requirement
		if ignoreGitRequirement || t.hasGitDirectory(basePath) {
			projectName := filepath.Base(basePath)
			project := ProjectInfo{
				Path:         basePath,
				Name:         projectName,
				LastModified: recentFileTime,
			}
			projects = append(projects, project)
			foundProjects[basePath] = true
		}
	}

	// Then, scan immediate subdirectories (only if base path wasn't already added as a project)
	entries, err = os.ReadDir(basePath)
	if err != nil {
		goto end
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		var projectPath string
		var isProject bool

		projectPath = filepath.Join(basePath, entry.Name())

		// Skip if we already found this path
		if foundProjects[projectPath] {
			continue
		}

		// Check if this directory qualifies as a project (subdirectories don't need 5+ files)
		isProject, err = t.isProjectDirectorySubdir(projectPath, ignoreGitRequirement)
		if err != nil {
			// Log error but continue
			logger.Error("Failed to check if directory is project", "path", projectPath, "error", err)
			continue
		}

		if !isProject {
			continue
		}

		// Get the most recent modification time from files in this directory
		recentFileTime, err = t.findMostRecentFileTime(projectPath)
		if err != nil {
			// Log error but continue - use directory modification time as fallback
			logger.Info("Failed to get recent file time", "path", projectPath, "error", err)
			var info os.FileInfo
			info, err = entry.Info()
			if err != nil {
				continue
			}
			recentFileTime = info.ModTime()
		}

		project := ProjectInfo{
			Path:         projectPath,
			Name:         entry.Name(),
			LastModified: recentFileTime,
		}

		projects = append(projects, project)
		foundProjects[projectPath] = true
	}

end:
	return projects, err
}

// isProjectDirectory checks if an allowed_path root is a project (requires 5+ files)
func (t *DetectCurrentProjectTool) isProjectDirectory(dirPath string, ignoreGitRequirement bool) (bool, error) {
	var fileCount int
	var err error

	// Check minimum file count threshold (5 files) - always required for allowed_path roots
	_, fileCount, err = t.findMostRecentFileTimeAndCount(dirPath)
	if err != nil || fileCount < 5 {
		return false, err
	}

	// Check git requirement
	if ignoreGitRequirement {
		// Check if this directory is inside a parent git repository (to avoid subdirectories)
		if t.isInsideGitRepository(dirPath) {
			return false, nil // It's inside a git repo, so not a separate project
		}
		return true, nil
	}

	// Otherwise, check for .git directory
	return t.hasGitDirectory(dirPath), nil
}

// isProjectDirectorySubdir checks if a subdirectory is a project (no file count requirement)
func (t *DetectCurrentProjectTool) isProjectDirectorySubdir(dirPath string, ignoreGitRequirement bool) (bool, error) {
	// Subdirectories don't need 5+ files, just git requirement check
	if ignoreGitRequirement {
		// Check if this directory is inside a parent git repository (to avoid subdirectories)
		if t.isInsideGitRepository(dirPath) {
			return false, nil // It's inside a git repo, so not a separate project
		}
		return true, nil
	}

	// Otherwise, check for .git directory
	return t.hasGitDirectory(dirPath), nil
}

func (t *DetectCurrentProjectTool) hasGitDirectory(dirPath string) bool {
	gitPath := filepath.Join(dirPath, ".git")
	if info, err := os.Stat(gitPath); err == nil {
		return info.IsDir()
	}
	return false
}

func (t *DetectCurrentProjectTool) isInsideGitRepository(dirPath string) bool {
	// Walk up the directory tree to see if we're inside a git repository
	currentPath := dirPath
	for {
		parent := filepath.Dir(currentPath)
		// If we've reached the root, stop
		if parent == currentPath {
			break
		}

		// Check if parent has .git directory
		if t.hasGitDirectory(parent) {
			return true
		}

		currentPath = parent
	}
	return false
}

func (t *DetectCurrentProjectTool) findMostRecentFileTime(dirPath string) (mostRecent time.Time, err error) {
	mostRecent, _, err = t.findMostRecentFileTimeAndCount(dirPath)
	return mostRecent, err
}

func (t *DetectCurrentProjectTool) findMostRecentFileTimeAndCount(dirPath string) (mostRecent time.Time, fileCount int, err error) {
	var entries []os.DirEntry

	entries, err = os.ReadDir(dirPath)
	if err != nil {
		goto end
	}

	mostRecent = time.Time{} // Zero time
	fileCount = 0

	for _, entry := range entries {
		// Skip hidden files and directories (including .git)
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		var info os.FileInfo
		filePath := filepath.Join(dirPath, entry.Name())

		info, err = entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		// For regular files, use modification time directly and count them
		if !info.IsDir() {
			fileCount++
			if info.ModTime().After(mostRecent) {
				mostRecent = info.ModTime()
			}
			continue
		}

		if !t.hasGitDirectory(filePath) {
			continue
		}

		if info.ModTime().After(mostRecent) {
			mostRecent = info.ModTime()
		}

	}
	// If no files found, return zero time and an error
	if fileCount == 0 {
		err = fmt.Errorf("no files found in directory")
	}

end:
	return mostRecent, fileCount, err
}

// excludeCurrentProject returns a slice of projects excluding the current project
func (t *DetectCurrentProjectTool) excludeCurrentProject(allProjects []ProjectInfo, currentProject *ProjectInfo) []ProjectInfo {
	// Initialize as empty slice to ensure JSON marshals as [] not null
	otherProjects := []ProjectInfo{}

	if currentProject == nil {
		return allProjects
	}

	for _, project := range allProjects {
		if project.Path != currentProject.Path {
			otherProjects = append(otherProjects, project)
		}
	}

	return otherProjects
}

func (t *DetectCurrentProjectTool) formatRelativeTime(modTime time.Time) string {
	now := time.Now()
	diff := now.Sub(modTime)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		months := int(diff.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
}
