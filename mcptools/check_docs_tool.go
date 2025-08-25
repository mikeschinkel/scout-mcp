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
				OffsetProperty,
			},
		}),
	})
}

// CheckDocsTool analyzes files and provides information about their structure and content.
type CheckDocsTool struct {
	*mcputil.ToolBase
}

const TargetCharLimit = 90000 // 90K safety margin as per PLAN.md

// Handle processes the check_docs tool request and returns documentation analysis results.
func (t *CheckDocsTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var recursive bool
	var exceptions []golang.DocException
	var path string
	var analysisResult *DocsAnalysisResult
	var language string
	var offset int

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

	offset, err = OffsetProperty.Int(req)
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

	// Apply intelligent response sizing and prioritization (includes offset handling)
	analysisResult = t.createSizedAnalysisResult(path, exceptions, offset)

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

type DocsAnalysisResultArgs struct {
	Path         string
	Exceptions   []golang.DocException
	TotalFound   int
	ResponseSize int
	Offset       int
}

func NewDocsAnalysisResult(args DocsAnalysisResultArgs) *DocsAnalysisResult {
	returnedCount := len(args.Exceptions)
	sizeLimited := args.TotalFound > returnedCount

	// Calculate remaining count considering offset
	// RemainingCount = issues not yet returned (total - offset - returned)
	remainingCount := args.TotalFound - args.Offset - returnedCount
	if remainingCount < 0 {
		remainingCount = 0
	}

	// Convert issues to use relative paths
	issues := NewDocsAnalysisIssues(args.Exceptions, args.Path)

	result := &DocsAnalysisResult{
		Path:           args.Path,
		Issues:         issues,
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
				"Run again with smaller max_files parameter or after fixing current issues.",
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

// createSizedAnalysisResult applies intelligent response sizing with prioritization
func (t *CheckDocsTool) createSizedAnalysisResult(path string, allExceptions []golang.DocException, offset int) (result *DocsAnalysisResult) {
	var exceptions []golang.DocException
	var maxIssues int
	var currentSize int
	var jsonBytes []byte

	totalCount := len(allExceptions)

	// Sort by priority first
	sortIssuesByPriority(allExceptions)

	// Apply offset for pagination
	exceptions = allExceptions
	if offset > 0 {
		if offset >= len(exceptions) {
			exceptions = []golang.DocException{}
		} else {
			exceptions = exceptions[offset:]
		}
	}

	// Dynamic size optimization - start with all available exceptions
	maxIssues = len(exceptions)

	// Iterative size reduction with ratio-based optimization
	for {
		result = NewDocsAnalysisResult(DocsAnalysisResultArgs{
			Path:         path,
			Exceptions:   exceptions[:maxIssues],
			TotalFound:   totalCount,
			ResponseSize: currentSize,
			Offset:       offset,
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
