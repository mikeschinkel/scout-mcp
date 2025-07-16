package mcptools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mikeschinkel/scout-mcp/langutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*ValidateFilesTool)(nil)

func init() {
	mcputil.RegisterTool(&ValidateFilesTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "validate_files",
			Description: "Validate syntax of source code files using language-specific parsers",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				FilesProperty,
				PathsProperty,
				LanguageProperty,
				RecursiveProperty,
				mcputil.Array("extensions", "Extensions of files to process for this tool"),
			},
		}),
	})
}

type ValidateFilesTool struct {
	*toolBase
}

type ValidationResult struct {
	FilePath string `json:"file_path"`
	Language string `json:"language"`
	Valid    bool   `json:"valid"`
	Error    string `json:"error,omitempty"`
}

type ValidationSummary struct {
	TotalFiles   int                `json:"total_files"`
	ValidFiles   int                `json:"valid_files"`
	InvalidFiles int                `json:"invalid_files"`
	Results      []ValidationResult `json:"results"`
	OverallValid bool               `json:"overall_valid"`
}

func (t *ValidateFilesTool) parseFilesOrPaths(req mcputil.ToolRequest) (files []string, paths []string, err error) {
	var filesParam []any
	var pathsParam []any
	// Parse parameters - files takes precedence over paths
	filesParam = req.GetArray("files", nil)
	if len(filesParam) > 0 {
		for _, f := range filesParam {
			if str, ok := f.(string); ok {
				files = append(files, str)
			}
		}
		goto end
	}
	// If no files specified, use paths
	pathsParam = req.GetArray("paths", nil)
	if len(pathsParam) > 0 {
		for _, p := range pathsParam {
			if str, ok := p.(string); ok {
				paths = append(paths, str)
			}
		}
		goto end
	}
	err = fmt.Errorf("either 'files' (array of file paths) or 'paths' (array of directory paths) parameter is required")
end:
	return files, paths, err
}
func (t *ValidateFilesTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var files []string
	var paths []string
	var languageOverride string
	var recursive bool
	var extensions []string
	var summary ValidationSummary
	var extensionsParam []any

	logger.Info("Tool called", "tool", "validate_files")

	files, paths, err = t.parseFilesOrPaths(req)
	if err != nil {
		goto end
	}

	languageOverride = req.GetString("language", "")
	recursive = req.GetBool("recursive", false)

	extensionsParam = req.GetArray("extensions", nil)
	for _, e := range extensionsParam {
		if str, ok := e.(string); ok {
			extensions = append(extensions, str)
		}
	}

	// Validate files
	if len(files) > 0 {
		summary, err = t.validateSpecificFiles(files, languageOverride)
	} else {
		summary, err = t.validateFilesByPaths(paths, languageOverride, recursive, extensions)
	}
	if err != nil {
		goto end
	}

	result = mcputil.NewToolResultJSON(summary)
	logger.Info("Tool completed", "tool", "validate_files", "total_files", summary.TotalFiles, "valid_files", summary.ValidFiles, "invalid_files", summary.InvalidFiles)

end:
	return result, err
}

func (t *ValidateFilesTool) validateSpecificFiles(files []string, languageOverride string) (summary ValidationSummary, err error) {
	var validCount int

	summary.TotalFiles = len(files)
	summary.Results = make([]ValidationResult, 0, len(files))

	// Validate each file
	for _, filePath := range files {
		var result ValidationResult
		result, err = t.validateSingleFile(filePath, languageOverride)
		if err != nil {
			goto end
		}

		summary.Results = append(summary.Results, result)
		if result.Valid {
			validCount++
		}
	}

	summary.ValidFiles = validCount
	summary.InvalidFiles = summary.TotalFiles - validCount
	summary.OverallValid = summary.InvalidFiles == 0

end:
	return summary, err
}

func (t *ValidateFilesTool) validateFilesByPaths(paths []string, languageOverride string, recursive bool, extensions []string) (summary ValidationSummary, err error) {
	var allFiles []string
	var validCount int

	// Collect all files to validate
	for _, path := range paths {
		var pathFiles []string
		pathFiles, err = t.collectFiles(path, recursive, extensions)
		if err != nil {
			goto end
		}
		allFiles = append(allFiles, pathFiles...)
	}

	summary.TotalFiles = len(allFiles)
	summary.Results = make([]ValidationResult, 0, len(allFiles))

	// Validate each file
	for _, filePath := range allFiles {
		var result ValidationResult
		result, err = t.validateSingleFile(filePath, languageOverride)
		if err != nil {
			goto end
		}

		summary.Results = append(summary.Results, result)
		if result.Valid {
			validCount++
		}
	}

	summary.ValidFiles = validCount
	summary.InvalidFiles = summary.TotalFiles - validCount
	summary.OverallValid = summary.InvalidFiles == 0

end:
	return summary, err
}

func (t *ValidateFilesTool) collectFiles(path string, recursive bool, extensions []string) (files []string, err error) {
	var allowed bool

	allowed, err = t.IsAllowedPath(path)
	if err != nil {
		goto end
	}

	if !allowed {
		err = fmt.Errorf("access denied: path not allowed: %s", path)
		goto end
	}

	// Use the search functionality from toolBase
	files, err = t.searchForFiles(path, recursive, extensions)

end:
	return files, err
}

func (t *ValidateFilesTool) validateSingleFile(filePath, languageOverride string) (result ValidationResult, err error) {
	var content string
	var language string

	result.FilePath = filePath

	// Read file content
	content, err = t.readFile(filePath)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("failed to read file: %v", err)
		// Don't return error - this is expected for unreadable files
		err = nil
		goto end
	}

	// Determine language
	if languageOverride != "" {
		language = languageOverride
	} else {
		language = t.detectLanguageFromExtension(filePath)
	}

	if language == "" {
		result.Valid = false
		result.Error = "could not determine language from file extension"
		goto end
	}

	result.Language = language

	// Validate syntax
	err = langutil.ValidateSyntax(language, content)
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
		// Don't return the validation error - it's expected for invalid files
		err = nil
	} else {
		result.Valid = true
	}

end:
	return result, err
}

func (t *ValidateFilesTool) detectLanguageFromExtension(filePath string) (language string) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		language = "go"
	case ".js", ".mjs":
		language = "javascript"
	case ".ts":
		language = "typescript"
	case ".py":
		language = "python"
	case ".rs":
		language = "rust"
	case ".java":
		language = "java"
	case ".c":
		language = "c"
	case ".cpp", ".cc", ".cxx":
		language = "cpp"
	default:
		language = ""
	}

	return language
}
