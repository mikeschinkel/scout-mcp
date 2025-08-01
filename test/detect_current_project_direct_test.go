package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectCurrentProjectDirect tests the detect_current_project functionality using direct server access
func TestDetectCurrentProjectDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("NoProjectDirectories", func(t *testing.T) {
		testNoProjectDirectoriesDirect(t, env)
	})

	t.Run("SingleProject", func(t *testing.T) {
		testSingleProjectDirect(t, env)
	})

	t.Run("MultipleProjectsClearWinner", func(t *testing.T) {
		testMultipleProjectsClearWinnerDirect(t, env)
	})

	t.Run("MultipleRecentProjects", func(t *testing.T) {
		testMultipleRecentProjectsDirect(t, env)
	})

	t.Run("ListRecentMode", func(t *testing.T) {
		testListRecentModeDirect(t, env)
	})

	t.Run("MaxProjectsLimit", func(t *testing.T) {
		testMaxProjectsLimitDirect(t, env)
	})

	t.Run("HiddenDirectoriesIgnored", func(t *testing.T) {
		testHiddenDirectoriesIgnoredDirect(t, env)
	})
}

func testNoProjectDirectoriesDirect(t *testing.T, env *DirectServerTestEnv) {
	cleanupTestDir(t, env.testDir)

	// Test when there are truly no project subdirectories and few files
	result := env.CallTool(t, "detect_current_project", map[string]interface{}{})

	var response map[string]interface{}
	ParseJSONResult(t, result, &response)

	// Should handle the case gracefully
	assert.Contains(t, response, "summary", "Response should contain summary")
	assert.Contains(t, response, "recent_projects", "Response should contain recent_projects")

	// With empty directory, should find no projects
	recentProjects := response["recent_projects"]
	if recentProjects != nil {
		recentProjectsList := recentProjects.([]interface{})
		assert.Equal(t, 0, len(recentProjectsList), "Should find no projects in empty directory")
	}

	summary := response["summary"].(string)
	assert.Contains(t, summary, "No projects found", "Summary should indicate no projects found")
}

func testSingleProjectDirect(t *testing.T, env *DirectServerTestEnv) {
	cleanupTestDir(t, env.testDir)

	// Create a single project directory with enough files (5+) and .git
	createTestProject(t, env.testDir, "my-awesome-project", true)

	result := env.CallTool(t, "detect_current_project", map[string]interface{}{})

	var response map[string]interface{}
	ParseJSONResult(t, result, &response)

	// Should identify the single project as current
	assert.Contains(t, response, "current_project", "Should have current_project")
	assert.Contains(t, response, "recent_projects", "Should have recent_projects")
	assert.Contains(t, response, "summary", "Should have summary")

	currentProject := response["current_project"].(map[string]interface{})
	assert.Equal(t, "my-awesome-project", currentProject["name"], "Should identify correct project name")

	summary := response["summary"].(string)
	assert.Contains(t, summary, "my-awesome-project", "Summary should mention project name")
	assert.Contains(t, summary, "only project found", "Should indicate it's the only project")
}

func testMultipleProjectsClearWinnerDirect(t *testing.T, env *DirectServerTestEnv) {
	cleanupTestDir(t, env.testDir)

	// Create multiple projects with clear time separation
	oldProjectDir := createTestProject(t, env.testDir, "old-project", true)
	recentProjectDir := createTestProject(t, env.testDir, "recent-project", true)

	// Set old project files to be more than 24 hours old
	oldTime := time.Now().Add(-25 * time.Hour)
	oldFiles := []string{"README.md", "main.go", "config.json", "package.json", "Makefile", "LICENSE"}
	for _, fileName := range oldFiles {
		oldFilePath := filepath.Join(oldProjectDir, fileName)
		err := os.Chtimes(oldFilePath, oldTime, oldTime)
		require.NoError(t, err, "Failed to set old project file time for %s", fileName)
	}

	// Make sure recent project has recent files by updating one of its files
	recentFile := filepath.Join(recentProjectDir, "README.md")
	err := os.WriteFile(recentFile, []byte("recently updated content"), 0644)
	require.NoError(t, err, "Failed to update recent project file")

	result := env.CallTool(t, "detect_current_project", map[string]interface{}{})

	var response map[string]interface{}
	ParseJSONResult(t, result, &response)

	// Should identify the recent project as current
	assert.Contains(t, response, "current_project", "Should have current_project")
	assert.False(t, response["requires_choice"].(bool), "Should not require user choice")

	currentProject := response["current_project"].(map[string]interface{})
	assert.Equal(t, "recent-project", currentProject["name"], "Should identify recent project as current")

	summary := response["summary"].(string)
	assert.Contains(t, summary, "recent-project", "Summary should mention current project")
}

func testMultipleRecentProjectsDirect(t *testing.T, env *DirectServerTestEnv) {
	cleanupTestDir(t, env.testDir)

	// Create multiple projects with recent times (within 24 hours)
	project1Dir := createTestProject(t, env.testDir, "project-alpha", true)
	project2Dir := createTestProject(t, env.testDir, "project-beta", true)

	// Set both to recent times (within 24 hours) by updating their files
	recentTime1 := time.Now().Add(-2 * time.Hour)
	alphaFile := filepath.Join(project1Dir, "README.md")
	err := os.Chtimes(alphaFile, recentTime1, recentTime1)
	require.NoError(t, err, "Failed to set project alpha file time")

	recentTime2 := time.Now().Add(-3 * time.Hour)
	betaFile := filepath.Join(project2Dir, "README.md")
	err = os.Chtimes(betaFile, recentTime2, recentTime2)
	require.NoError(t, err, "Failed to set project beta file time")

	result := env.CallTool(t, "detect_current_project", map[string]interface{}{})

	var response map[string]interface{}
	ParseJSONResult(t, result, &response)

	// Should require user choice since multiple projects are recent
	assert.True(t, response["requires_choice"].(bool), "Should require user choice")
	assert.Contains(t, response, "choice_message", "Should have choice message")
	assert.Contains(t, response, "recent_projects", "Should have recent projects list")

	recentProjects := response["recent_projects"].([]interface{})
	assert.GreaterOrEqual(t, len(recentProjects), 2, "Should have at least 2 recent projects")

	choiceMessage := response["choice_message"].(string)
	assert.Contains(t, choiceMessage, "Multiple projects modified recently", "Should explain the situation")
}

func testListRecentModeDirect(t *testing.T, env *DirectServerTestEnv) {
	cleanupTestDir(t, env.testDir)

	// Create several projects with different modification times
	for i := 0; i < 7; i++ {
		projectDir := createTestProject(t, env.testDir, fmt.Sprintf("project-%d", i), true)

		// Set different modification times by updating files
		modTime := time.Now().Add(-time.Duration(i) * time.Hour)
		readmeFile := filepath.Join(projectDir, "README.md")
		err := os.Chtimes(readmeFile, modTime, modTime)
		require.NoError(t, err, "Failed to set project %d file time", i)
	}

	result := env.CallTool(t, "detect_current_project", map[string]interface{}{
		"list_recent":  true,
		"max_projects": 5,
	})

	var response map[string]interface{}
	ParseJSONResult(t, result, &response)

	// In list mode, should not detect current project but list recent ones
	_, hasCurrentProject := response["current_project"]
	assert.False(t, hasCurrentProject, "Should not have current_project in list mode")

	assert.Contains(t, response, "recent_projects", "Should have recent_projects")
	recentProjects := response["recent_projects"].([]interface{})
	assert.LessOrEqual(t, len(recentProjects), 5, "Should respect max_projects limit")
	assert.GreaterOrEqual(t, len(recentProjects), 1, "Should have at least one project")

	// Check that projects have required fields
	for i, projectInterface := range recentProjects {
		project := projectInterface.(map[string]interface{})
		assert.Contains(t, project, "name", "Project %d should have name", i)
		assert.Contains(t, project, "path", "Project %d should have path", i)
		assert.Contains(t, project, "last_modified", "Project %d should have last_modified", i)
		assert.Contains(t, project, "relative_age", "Project %d should have relative_age", i)

		// Verify relative_age is human-readable
		relativeAge := project["relative_age"].(string)
		assert.NotEmpty(t, relativeAge, "Relative age should not be empty")
		assert.True(t,
			contains(relativeAge, "hour") || contains(relativeAge, "minute") || contains(relativeAge, "day") || contains(relativeAge, "just now"),
			"Relative age should be human-readable: %s", relativeAge)
	}
}

func testMaxProjectsLimitDirect(t *testing.T, env *DirectServerTestEnv) {
	cleanupTestDir(t, env.testDir)

	// Create more projects than the limit
	for i := 0; i < 8; i++ {
		projectDir := createTestProject(t, env.testDir, fmt.Sprintf("project-limit-%d", i), true)

		// Set different modification times by updating files
		modTime := time.Now().Add(-time.Duration(i) * time.Hour)
		readmeFile := filepath.Join(projectDir, "README.md")
		err := os.Chtimes(readmeFile, modTime, modTime)
		require.NoError(t, err, "Failed to set project %d file time", i)
	}

	result := env.CallTool(t, "detect_current_project", map[string]interface{}{
		"max_projects": 3,
	})

	var response map[string]interface{}
	ParseJSONResult(t, result, &response)

	assert.Contains(t, response, "recent_projects", "Should have recent_projects")
	recentProjects := response["recent_projects"].([]interface{})
	assert.LessOrEqual(t, len(recentProjects), 3, "Should respect max_projects limit of 3")
}

func testHiddenDirectoriesIgnoredDirect(t *testing.T, env *DirectServerTestEnv) {
	cleanupTestDir(t, env.testDir)

	// Create hidden directories (should be ignored)
	hiddenDir := filepath.Join(env.testDir, ".hidden-project")
	err := os.Mkdir(hiddenDir, 0755)
	require.NoError(t, err, "Failed to create hidden directory")

	// Create normal project directory with enough files (but without .git in parent)
	createTestProject(t, env.testDir, "normal-project", false) // No .git for this project

	// Create another project directory with .git (this should be found)
	createTestProject(t, env.testDir, "git-project", true) // With .git

	result := env.CallTool(t, "detect_current_project", map[string]interface{}{
		"ignore_git_requirement": true, // This test focuses on hidden directory filtering, not git detection
	})

	var response map[string]interface{}
	ParseJSONResult(t, result, &response)

	// Should find visible projects but not hidden ones
	assert.Contains(t, response, "recent_projects", "Should have recent_projects")

	recentProjects := response["recent_projects"].([]interface{})
	assert.GreaterOrEqual(t, len(recentProjects), 1, "Should find at least one visible project")

	// Verify no hidden directories are included in the results
	projectNames := []string{}
	for _, projectInterface := range recentProjects {
		project := projectInterface.(map[string]interface{})
		projectNames = append(projectNames, project["name"].(string))
	}

	// Should contain normal projects but not hidden ones
	assert.Contains(t, projectNames, "normal-project", "Should find normal project")
	assert.Contains(t, projectNames, "git-project", "Should find git project")
	assert.NotContains(t, projectNames, ".hidden-project", "Should not find hidden project")
}

// Helper function to create a project directory with enough files and .git
func createTestProject(t *testing.T, baseDir, projectName string, withGit bool) string {
	projectDir := filepath.Join(baseDir, projectName)
	err := os.Mkdir(projectDir, 0755)
	require.NoError(t, err, "Failed to create project directory %s", projectName)

	if withGit {
		// Create .git directory to make it a valid project
		gitDir := filepath.Join(projectDir, ".git")
		err = os.Mkdir(gitDir, 0755)
		require.NoError(t, err, "Failed to create .git directory for %s", projectName)
	}

	// Create enough files (5+) to meet threshold
	testFiles := []string{"README.md", "main.go", "config.json", "package.json", "Makefile", "LICENSE"}
	for _, fileName := range testFiles {
		filePath := filepath.Join(projectDir, fileName)
		err = os.WriteFile(filePath, []byte("test content for "+projectName), 0644)
		require.NoError(t, err, "Failed to create test file %s in %s", fileName, projectName)
	}

	return projectDir
}

// Helper function to clean up test directory
func cleanupTestDir(t *testing.T, testDir string) {
	t.Helper()
	entries, _ := os.ReadDir(testDir)
	for _, entry := range entries {
		entryPath := filepath.Join(testDir, entry.Name())
		if entry.IsDir() {
			must(os.RemoveAll(entryPath))
		} else {
			must(os.Remove(entryPath))
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
