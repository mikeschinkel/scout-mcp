package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

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
				RecursiveProperty,
				OffsetProperty,
			},
		}),
	})
}

// CheckDocsTool analyzes files and provides information about their structure and content.
type CheckDocsTool struct {
	*mcputil.ToolBase
}

const TARGET_CHAR_LIMIT = 90000 // 90K safety margin as per PLAN.md

// Handle processes the check_docs tool request and returns documentation analysis results.
func (t *CheckDocsTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var recursive bool
	var allExceptions []golang.DocException
	var path string
	var analysisResult *DocsAnalysisResult

	logger.Info("Tool called", "tool", t.Name())

	path, err = RequiredPathProperty.String(req)
	if err != nil {
		goto end
	}

	recursive, err = RecursiveProperty.Bool(req)
	if err != nil {
		goto end
	}

	// Get all documentation exceptions
	allExceptions, err = golang.DocExceptions(context.Background(), &golang.DocsExceptionsArgs{
		Path:      path,
		Recursive: golang.GetRecurseDirective(recursive),
	})
	if err != nil {
		goto end
	}

	// Apply intelligent response sizing and prioritization
	analysisResult = t.createSizedAnalysisResult(path, allExceptions)

	logger.Info("Tool completed", "tool", t.Name(),
		"total_issues", analysisResult.TotalCount,
		"returned_issues", analysisResult.ReturnedCount,
		"size_limited", analysisResult.SizeLimited,
		"response_size", analysisResult.ResponseSize)

	result = mcputil.NewToolResultJSON(analysisResult)

end:
	return result, err
}

type DocsAnalysisResult struct {
	Path           string              `json:"path"`
	Issues         []DocsAnalysisIssue `json:"issues"`
	ReturnedCount  int                 `json:"returned_count"`
	TotalCount     int                 `json:"total_count"`
	RemainingCount int                 `json:"remaining_count"`
	SizeLimited    bool                `json:"size_limited"`
	ResponseSize   int                 `json:"response_size_chars"`
	Message        string              `json:"message,omitempty"`
}

func NewDocsAnalysisResult(path string, exceptions []golang.DocException, totalFound int, responseSize int) *DocsAnalysisResult {
	returnedCount := len(exceptions)
	sizeLimited := totalFound > returnedCount

	// Convert issues to use relative paths
	issues := NewDocsAnalysisIssues(exceptions, path)

	result := &DocsAnalysisResult{
		Path:           path,
		Issues:         issues,
		ReturnedCount:  returnedCount,
		TotalCount:     totalFound,
		RemainingCount: totalFound - returnedCount,
		SizeLimited:    sizeLimited,
		ResponseSize:   responseSize,
	}

	if sizeLimited {
		result.Message = fmt.Sprintf(
			"Response limited to %d of %d total issues due to size constraints (%d chars). "+
				"Showing highest priority issues first. %d issues remaining. "+
				"Run again with smaller max_files parameter or after fixing current issues.",
			returnedCount, totalFound, responseSize, result.RemainingCount)
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

// createSizedAnalysisResult applies intelligent response sizing with prioritization
func (t *CheckDocsTool) createSizedAnalysisResult(path string, allExceptions []golang.DocException) *DocsAnalysisResult {
	totalCount := len(allExceptions)
	if totalCount == 0 {
		return NewDocsAnalysisResult(path, allExceptions, totalCount, 0)
	}

	// Sort by priority first
	sortIssuesByPriority(allExceptions)

	// Apply user-specified limits if provided
	exceptions := allExceptions

	// Dynamic size optimization
	maxIssues := len(exceptions)
	testResult := NewDocsAnalysisResult(path, exceptions[:maxIssues], totalCount, 0)
	jsonBytes, _ := json.Marshal(testResult)
	currentSize := len(jsonBytes)

	// Iterative size reduction with ratio-based optimization
	for currentSize > TARGET_CHAR_LIMIT && maxIssues > 1 {
		// Calculate ratio for efficient reduction
		ratio := float64(TARGET_CHAR_LIMIT) / float64(currentSize)
		newMaxIssues := int(float64(maxIssues) * ratio * 0.95) // 5% buffer

		// Ensure progress
		if newMaxIssues >= maxIssues {
			newMaxIssues = maxIssues - 1
		}
		if newMaxIssues < 1 {
			newMaxIssues = 1
		}

		maxIssues = newMaxIssues

		// Test new size
		truncatedExceptions := exceptions[:maxIssues]
		testResult = NewDocsAnalysisResult(path, truncatedExceptions, totalCount, 0)
		jsonBytes, _ = json.Marshal(testResult)
		currentSize = len(jsonBytes)
	}

	// Create final result with actual size
	finalResult := NewDocsAnalysisResult(path, exceptions[:maxIssues], totalCount, currentSize)
	return finalResult
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
