package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mikeschinkel/scout-mcp/langutil"
	"github.com/mikeschinkel/scout-mcp/langutil/golang"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*CheckDocsTool)(nil)

func init() {
	mcputil.RegisterTool(&CheckDocsTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name: "check_docs",
			// TODO: Add a better description
			Description: "Check Documentation",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				RequiredPathProperty,
				RequiredLanguageProperty,
				RecursiveProperty,
			},
		}),
	})
}

// CheckDocsTool analyzes files and provides information about their structure and content.
type CheckDocsTool struct {
	*mcputil.ToolBase
}

const TargetCharLimit = 75000 // 75K provided safety margin for 25k tokens max

// Handle processes the check_docs tool request and returns documentation analysis results.
func (t *CheckDocsTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var recursive bool
	var exceptions []golang.DocException
	var path string
	var analysisResult *DocsAnalysisResult
	var language string

	logger.Info("Tool called", "tool", t.Name())

	language, err = LanguageProperty.String(req)
	if err != nil {
		goto end
	}
	if language != string(langutil.GoLanguage) {
		err = fmt.Errorf("the '%s' language not currently (yet?) supported by 'check_docs' tool", language)
		goto end
	}

	path, err = RequiredPathProperty.String(req)
	if err != nil {
		goto end
	}

	recursive, err = RecursiveProperty.Bool(req)
	if err != nil {
		goto end
	}

	// Get all documentation exceptions (without offset first)
	exceptions, err = golang.DocExceptions(context.Background(), &golang.DocsExceptionsArgs{
		Path:      path,
		Recursive: golang.GetRecurseDirective(recursive),
	})
	if err != nil {
		goto end
	}

	// Apply intelligent response sizing and prioritization
	analysisResult = t.createSizedAnalysisResult(path, exceptions)

	logger.Info("Tool completed", "tool", t.Name(),
		"language", language,
		"total_issues", analysisResult.TotalCount,
		"returned_issues", analysisResult.ReturnedCount,
		"size_limited", analysisResult.SizeLimited,
		"response_size", analysisResult.ResponseSize)

	result = mcputil.NewToolResultJSON(analysisResult)

end:
	return result, err
}

type FileIssueGroup struct {
	File       string              `json:"file"`
	IssueCount int                 `json:"issue_count"`
	Issues     []DocsAnalysisIssue `json:"issues"`
}

type FileIssueCountItem struct {
	File       string `json:"file"`
	IssueCount int    `json:"issue_count"`
}

type IssueSummary struct {
	TotalFilesWithIssues int                  `json:"total_files_with_issues"`
	TotalIssues          int                  `json:"total_issues"`
	FilesByIssueCount    []FileIssueCountItem `json:"files_by_issue_count"`
}

type DocsAnalysisResult struct {
	Path           string           `json:"path"`
	IssuesByFile   []FileIssueGroup `json:"issues_by_file"`
	Summary        IssueSummary     `json:"summary"`
	ReturnedCount  int              `json:"returned_count"`
	TotalCount     int              `json:"total_count"`
	RemainingCount int              `json:"remaining_count"`
	SizeLimited    bool             `json:"size_limited"`
	ResponseSize   int              `json:"response_size_chars"`
	Message        string           `json:"message,omitempty"`
}

type DocsAnalysisResultArgs struct {
	Path         string
	Exceptions   []golang.DocException
	TotalFound   int
	ResponseSize int
}

func NewDocsAnalysisResult(args DocsAnalysisResultArgs) (result *DocsAnalysisResult) {
	var fileGroups []FileIssueGroup
	var returnedCount int
	var sizeLimited bool
	var remainingCount int
	var issues []DocsAnalysisIssue
	var summary IssueSummary
	var filesByIssueCount []FileIssueCountItem

	// Convert flat issues to grouped structure
	issues = NewDocsAnalysisIssues(args.Exceptions, args.Path)
	fileGroups = groupIssuesByFile(issues)

	returnedCount = len(args.Exceptions)
	sizeLimited = args.TotalFound > returnedCount

	// Calculate remaining count (simple calculation without offset)
	remainingCount = args.TotalFound - returnedCount
	if remainingCount < 0 {
		remainingCount = 0
	}

	// Create summary with files ordered by issue count (descending)
	filesByIssueCount = make([]FileIssueCountItem, 0, len(fileGroups))
	for _, group := range fileGroups {
		filesByIssueCount = append(filesByIssueCount, FileIssueCountItem{
			File:       group.File,
			IssueCount: group.IssueCount,
		})
	}

	// Sort files by issue count (descending)
	sort.Slice(filesByIssueCount, func(i, j int) bool {
		return filesByIssueCount[i].IssueCount > filesByIssueCount[j].IssueCount
	})

	summary = IssueSummary{
		TotalFilesWithIssues: len(fileGroups),
		TotalIssues:          returnedCount,
		FilesByIssueCount:    filesByIssueCount,
	}

	result = &DocsAnalysisResult{
		Path:           args.Path,
		IssuesByFile:   fileGroups,
		Summary:        summary,
		ReturnedCount:  returnedCount,
		TotalCount:     args.TotalFound,
		RemainingCount: remainingCount,
		SizeLimited:    sizeLimited,
		ResponseSize:   args.ResponseSize,
	}

	if sizeLimited {
		result.Message = fmt.Sprintf(
			"Response limited to %d of %d total issues due to size constraints (%d chars). "+
				"Showing highest priority issues first. %d issues remaining. "+
				"Run again after fixing current issues to see remaining ones.",
			returnedCount, args.TotalFound, args.ResponseSize, result.RemainingCount)
	}

	return result
}

func NewDocsAnalysisIssues(exceptions []golang.DocException, basePath string) (issues []DocsAnalysisIssue) {
	issues = make([]DocsAnalysisIssue, len(exceptions))
	for i, exception := range exceptions {
		issues[i] = NewDocsAnalysisIssue(exception, basePath)
	}
	return issues
}

type DocsAnalysisIssue struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	EndLine   *int   `json:"end_line,omitempty"`
	Issue     string `json:"issue"`
	Element   string `json:"element"`
	MultiLine bool   `json:"multi_line"`
}

func NewDocsAnalysisIssue(e golang.DocException, basePath string) (r DocsAnalysisIssue) {
	// Convert absolute path to relative path
	relativePath := e.File
	if strings.HasPrefix(e.File, basePath) {
		if rel, err := filepath.Rel(basePath, e.File); err == nil {
			relativePath = rel
		}
	}

	return DocsAnalysisIssue{
		File:      relativePath, // Now relative to basePath
		Line:      e.Line,
		EndLine:   e.EndLine,
		Issue:     e.Issue(),
		Element:   e.Element,
		MultiLine: e.MultiLine,
	}
}

func groupIssuesByFile(issues []DocsAnalysisIssue) (fileGroups []FileIssueGroup) {
	var fileGroupMap map[string][]DocsAnalysisIssue
	var group []DocsAnalysisIssue
	var exists bool
	var seenFiles []string

	fileGroupMap = make(map[string][]DocsAnalysisIssue)

	// Group issues by file, preserving order
	for _, issue := range issues {
		group, exists = fileGroupMap[issue.File]
		if !exists {
			seenFiles = append(seenFiles, issue.File)
		}
		group = append(group, issue)
		fileGroupMap[issue.File] = group
	}

	// Convert to array format, maintaining file priority order
	fileGroups = make([]FileIssueGroup, 0, len(seenFiles))
	for _, filePath := range seenFiles {
		group = fileGroupMap[filePath]
		fileGroups = append(fileGroups, FileIssueGroup{
			File:       filePath,
			IssueCount: len(group),
			Issues:     group,
		})
	}

	return fileGroups
}

// createSizedAnalysisResult applies intelligent response sizing with prioritization
func (t *CheckDocsTool) createSizedAnalysisResult(path string, allExceptions []golang.DocException) (result *DocsAnalysisResult) {
	var exceptions []golang.DocException
	var maxIssues int
	var currentSize int
	var jsonBytes []byte

	totalCount := len(allExceptions)

	// Sort by priority first
	sortIssuesByPriority(allExceptions)

	// Start with all exceptions (no offset)
	exceptions = allExceptions

	// Dynamic size optimization - start with all available exceptions
	maxIssues = len(exceptions)

	// Iterative size reduction with ratio-based optimization
	for {
		result = NewDocsAnalysisResult(DocsAnalysisResultArgs{
			Path:         path,
			Exceptions:   exceptions[:maxIssues],
			TotalFound:   totalCount,
			ResponseSize: currentSize,
		})
		// Handle empty case
		if totalCount == 0 {
			exceptions = allExceptions
			goto end
		}
		if maxIssues == 0 {
			goto end
		}

		// Calculate initial size
		jsonBytes, _ = json.Marshal(result)
		currentSize = len(jsonBytes)
		if currentSize < TargetCharLimit {
			goto end
		}
		// Calculate ratio for efficient reduction
		ratio := float64(TargetCharLimit) / float64(currentSize)
		newMaxIssues := int(float64(maxIssues) * ratio * 0.95) // 5% buffer

		// Ensure progress
		if newMaxIssues >= maxIssues {
			newMaxIssues = maxIssues - 1
		}
		if newMaxIssues < 1 {
			newMaxIssues = 1
		}

		maxIssues = newMaxIssues

		logger.Info("Size limiting iteration",
			"maxIssues", maxIssues,
			"currentSize", currentSize,
			"targetLimit", TargetCharLimit,
			"ratio", ratio)
	}

end:
	return result
}

// getIssueSeverity returns priority level for different exception types
func getIssueSeverity(docType golang.DocExceptionType) int {
	switch docType {
	case golang.FuncException, golang.TypeException,
		golang.ConstException, golang.VarException,
		golang.GroupException:
		return 1 // High priority
	case golang.FileException:
		return 2 // Medium priority
	case golang.ReadmeException:
		return 3 // Low priority
	default:
		return 4 // Lowest priority
	}
}

// sortIssuesByPriority sorts exceptions by priority, then by file path for consistency
func sortIssuesByPriority(exceptions []golang.DocException) {
	sort.Slice(exceptions, func(i, j int) bool {
		severityI := getIssueSeverity(exceptions[i].Type)
		severityJ := getIssueSeverity(exceptions[j].Type)

		if severityI != severityJ {
			return severityI < severityJ // Lower number = higher priority
		}

		// Secondary sort by file path for consistent ordering
		if exceptions[i].File != exceptions[j].File {
			return exceptions[i].File < exceptions[j].File
		}

		// Tertiary sort by line number
		return exceptions[i].Line < exceptions[j].Line
	})
}
