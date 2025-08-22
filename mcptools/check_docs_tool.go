package mcptools

import (
	"context"

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
				MaxFilesProperty,
			},
		}),
	})
}

// CheckDocsTool analyzes files and provides information about their structure and content.
type CheckDocsTool struct {
	*mcputil.ToolBase
}

// Handle processes the analyze_files tool request and returns file analysis results.
func (t *CheckDocsTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var recursive bool
	var maxFiles int
	var exceptions []golang.DocException
	var path string

	logger.Info("Tool called", "tool", t.Name())

	path, err = RequiredPathProperty.String(req)
	if err != nil {
		goto end
	}

	recursive, err = RecursiveProperty.Bool(req)
	if err != nil {
		goto end
	}

	maxFiles, err = MaxFilesProperty.Int(req)
	if err != nil {
		goto end
	}

	exceptions, err = golang.DocExceptions(context.Background(), &golang.DocsExceptionsArgs{
		Path:      path,
		Recursive: golang.GetRecurseDirective(recursive),
		MaxFiles:  maxFiles,
	})
	if err != nil {
		goto end
	}

	logger.Info("Tool completed", "tool", t.Name())

	result = mcputil.NewToolResultJSON(NewDocsAnalysisResult(path, exceptions))

end:
	return result, err
}

type DocsAnalysisResult struct {
	Path   string              `json:"path"`
	Issues []DocsAnalysisIssue `json:"issues"`
	Total  int                 `json:"total"`
}

func NewDocsAnalysisResult(path string, exceptions []golang.DocException) (r *DocsAnalysisResult) {
	return &DocsAnalysisResult{
		Path:   path,
		Issues: NewDocsAnalysisIssues(exceptions),
		Total:  len(exceptions),
	}
}
func NewDocsAnalysisIssues(exceptions []golang.DocException) (issues []DocsAnalysisIssue) {
	issues = make([]DocsAnalysisIssue, len(exceptions))
	for i, exception := range exceptions {
		issues[i] = NewDocsAnalysisIssue(exception)
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

func NewDocsAnalysisIssue(e golang.DocException) (r DocsAnalysisIssue) {
	return DocsAnalysisIssue{
		File:      e.File,
		Line:      e.Line,
		EndLine:   e.EndLine,
		Issue:     e.Issue(),
		Element:   e.Element,
		MultiLine: e.MultiLine,
	}
}
