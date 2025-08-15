package mcptools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*AnalyzeFilesTool)(nil)

func init() {
	mcputil.RegisterTool(&AnalyzeFilesTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name: "analyze_files",
			// TODO: Add a better description
			Description: "Analyze files",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				RequiredFilesProperty,
			},
		}),
	})
}

type AnalyzeFilesTool struct {
	*mcputil.ToolBase
}

func (t *AnalyzeFilesTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var files []string
	var fileResults []FileAnalysisResult
	var totalFiles, totalErrors int
	var totalSize int64

	logger.Info("Tool called", "tool", "analyze_files")

	files, err = FilesProperty.StringSlice(req)
	if err != nil {
		goto end
	}

	if len(files) == 0 {
		err = fmt.Errorf("no files to analyze; the 'files' parameter cannot be empty")
		goto end
	}

	fileResults, totalSize, totalErrors, err = t.analyzeFiles(files)
	if err != nil {
		goto end
	}

	totalFiles = len(fileResults)
	logger.Info("Tool completed", "tool", "analyze_files", "files_analyzed", totalFiles, "total_size", totalSize, "errors", totalErrors)

	result = mcputil.NewToolResultJSON(map[string]any{
		"files":        fileResults,
		"total_files":  totalFiles,
		"total_size":   totalSize,
		"total_errors": totalErrors,
		"summary":      fmt.Sprintf("Analyzed %d files (%d bytes total)", totalFiles, totalSize),
	})

end:
	return result, err
}

type FileAnalysisResult struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Lines int    `json:"lines,omitempty"`
	Error string `json:"error,omitempty"`
}

func (t *AnalyzeFilesTool) analyzeFiles(filePaths []string) (results []FileAnalysisResult, totalSize int64, totalErrors int, err error) {
	results = make([]FileAnalysisResult, 0, len(filePaths))

	for _, filePath := range filePaths {
		var result FileAnalysisResult
		var fileInfo os.FileInfo
		var content []byte

		// Check path access
		if !t.IsAllowedPath(filePath) {
			result = FileAnalysisResult{
				Path:  filePath,
				Name:  filepath.Base(filePath),
				Error: "access denied: path not allowed",
			}
			results = append(results, result)
			totalErrors++
			continue
		}

		// Get file info
		fileInfo, err = os.Stat(filePath)
		if err != nil {
			result = FileAnalysisResult{
				Path:  filePath,
				Name:  filepath.Base(filePath),
				Error: fmt.Sprintf("cannot stat file: %v", err),
			}
			results = append(results, result)
			totalErrors++
			err = nil // Continue with other files
			continue
		}

		// Read file content to count lines
		content, err = os.ReadFile(filePath)
		if err != nil {
			result = FileAnalysisResult{
				Path:  filePath,
				Name:  filepath.Base(filePath),
				Size:  fileInfo.Size(),
				Error: fmt.Sprintf("cannot read file: %v", err),
			}
			results = append(results, result)
			totalErrors++
			err = nil // Continue with other files
			continue
		}

		// Count lines
		lines := strings.Count(string(content), "\n") + 1
		if len(content) == 0 {
			lines = 0
		}

		result = FileAnalysisResult{
			Path:  filePath,
			Name:  filepath.Base(filePath),
			Size:  fileInfo.Size(),
			Lines: lines,
		}

		results = append(results, result)
		totalSize += fileInfo.Size()
	}

	return results, totalSize, totalErrors, err
}
