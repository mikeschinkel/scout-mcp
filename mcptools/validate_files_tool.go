package mcptools

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/scout-mcp/fileutil"
	"github.com/mikeschinkel/scout-mcp/langutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*ValidateFilesTool)(nil)

func init() {
	mcputil.RegisterTool(&ValidateFilesTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "validate_files",
			Description: "Validate syntax of source code files using language-specific parsers",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				FilesProperty,
				PathsProperty,
				LanguageProperty,
				RecursiveProperty,
				ExtensionsProperty.Description("Extensions of files to process for this tool"),
			},
			Requires: []mcputil.Requirement{
				mcputil.RequiresOneOf{
					ParamNames: []string{"files", "paths"},
					Message:    "Either 'files' (array of file paths) or 'paths' (array of directory paths) parameter is required",
				},
			},
		}),
	})
}

// ValidateFilesTool validates syntax of source code files using language-specific parsers.
type ValidateFilesTool struct {
	*mcputil.ToolBase
}

type ValidationResult struct {
	FilePath string            `json:"file_path"`
	Language langutil.Language `json:"language"`
	Valid    bool              `json:"valid"`
	Error    string            `json:"error,omitempty"`
}

type ValidationSummary struct {
	TotalFiles   int                `json:"total_files"`
	ValidFiles   int                `json:"valid_files"`
	InvalidFiles int                `json:"invalid_files"`
	Results      []ValidationResult `json:"results"`
	OverallValid bool               `json:"overall_valid"`
}

func (t *ValidateFilesTool) parseFilesOrPaths(req mcputil.ToolRequest) (files []string, paths []string, err error) {
	var hasFiles, hasPaths bool

	// Parse parameters - files takes precedence over paths
	files, err = FilesProperty.StringSlice(req)
	if err == nil && len(files) > 0 {
		hasFiles = true
	}

	// If no files specified, use paths
	paths, err = PathsProperty.StringSlice(req)
	if err != nil || len(paths) > 0 {
		hasPaths = true
	}
	switch {
	case hasFiles && hasPaths:
		err = fmt.Errorf("cannot have both 'files' (array of file paths) and 'paths' (array of directory paths) parameters; provide only one")
	case !hasFiles && !hasPaths:
		err = fmt.Errorf("must have either a 'files' (array of file paths) or a 'paths' (array of directory paths) parameter")
	}

	return files, paths, err
}

// Handle processes the validate_files tool request and performs syntax validation on source files.
func (t *ValidateFilesTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var files []string
	var language string
	var results []langutil.ValidationResult
	var summary ValidationSummary
	var ffArgs fileutil.FindFileArgs

	logger.Info("Tool called", "tool", "validate_files")

	files, ffArgs.Paths, err = t.parseFilesOrPaths(req)
	if err != nil {
		goto end
	}

	ffArgs.Recursive, err = RecursiveProperty.Bool(req)
	if err != nil {
		goto end
	}

	ffArgs.Extensions, err = ExtensionsProperty.StringSlice(req)
	if err != nil {
		goto end
	}

	language, err = LanguageProperty.String(req)
	if err != nil {
		goto end
	}

	// Validate files
	if len(ffArgs.Paths) > 0 {
		files, err = fileutil.FindFiles(ffArgs)
	}

	// Errors returned ValidateFilesAs by SHOULD be ignored.
	// Teh MCP Server should get errors as information, not as an error
	results, _ = langutil.ValidateFilesAs(files, langutil.Language(language))
	summary = generateValidationSummary(results)
	result = mcputil.NewToolResultJSON(summary)
	logger.Info("Tool completed", "tool", "validate_files", "total_files", summary.TotalFiles, "valid_files", summary.ValidFiles, "invalid_files", summary.InvalidFiles)

end:
	return result, err
}

func generateValidationSummary(results []langutil.ValidationResult) (summary ValidationSummary) {

	summary.TotalFiles = len(results)
	summary.Results = make([]ValidationResult, 0, len(results))

	for _, result := range results {
		summary.Results = append(summary.Results, ValidationResult{
			FilePath: result.FilePath,
			Language: result.Language,
			Valid:    result.Error == nil,
			Error: func() (err string) {
				if result.Error != nil {
					err = result.Error.Error()
				}
				return err
			}(),
		})
		if result.Error == nil {
			summary.ValidFiles++
		}
	}
	summary.InvalidFiles = summary.TotalFiles - summary.ValidFiles
	summary.OverallValid = summary.InvalidFiles == 0

	return summary
}
