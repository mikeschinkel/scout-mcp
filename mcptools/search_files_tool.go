package mcptools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*SearchFilesTool)(nil)

type FileSearchResult struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	IsDir    bool   `json:"is_directory"`
}

func init() {
	mcputil.RegisterTool(&SearchFilesTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "search_files",
			Description: "Search for files and directories in allowed paths with filtering options",
			QuickHelp:   "Find files matching criteria",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				PathProperty.Required(),
				RecursiveProperty,
				ExtensionsProperty,
				PatternProperty.Description("Name pattern to match (case-insensitive substring)"),
				NamePatternProperty,
				FilesOnlyProperty,
				DirsOnlyProperty,
				MaxResultsProperty,
			},
		}),
	})
}

type SearchFilesTool struct {
	*toolBase
}

func (t *SearchFilesTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var searchPath string
	var recursive bool
	var pattern string
	var namePattern string
	var filesOnly bool
	var dirsOnly bool
	var maxResults int
	var extensions []string
	var results []FileSearchResult

	logger.Info("Tool called", "tool", "search_files")

	searchPath, err = PathProperty.String(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	recursive, err = RecursiveProperty.Bool(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	pattern, err = PatternProperty.String(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	namePattern, err = NamePatternProperty.String(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	filesOnly, err = FilesOnlyProperty.Bool(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	dirsOnly, err = DirsOnlyProperty.Bool(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	maxResults, err = MaxResultsProperty.Int(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	extensions, err = ExtensionsProperty.StringSlice(req)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("invalid extensions array: %v", err))
		goto end
	}

	logger.Info("Tool arguments parsed",
		"tool", "search_files",
		"path", searchPath,
		"recursive", recursive,
		"pattern", pattern,
		"name_pattern", namePattern,
		"files_only", filesOnly,
		"dirs_only", dirsOnly,
		"extensions", extensions,
		"max_results", maxResults)

	// Check path is allowed
	if !t.IsAllowedPath(searchPath) {
		result = mcputil.NewToolResultError(fmt.Errorf("access denied: path not allowed: %s", searchPath))
		goto end
	}

	results, err = t.searchFiles(searchPath, SearchFilesOptions{
		Recursive:   recursive,
		Pattern:     pattern,
		NamePattern: namePattern,
		Extensions:  extensions,
		FilesOnly:   filesOnly,
		DirsOnly:    dirsOnly,
		MaxResults:  maxResults,
	})
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	logger.Info("Tool completed", "tool", "search_files", "results_count", len(results))

	// Convert results to JSON using mcputil
	result = mcputil.NewToolResultJSON(map[string]any{
		"search_path":  searchPath,
		"results":      results,
		"count":        len(results),
		"recursive":    recursive,
		"pattern":      pattern,
		"name_pattern": namePattern,
		"extensions":   extensions,
		"files_only":   filesOnly,
		"dirs_only":    dirsOnly,
		"max_results":  maxResults,
		"truncated":    len(results) >= maxResults,
	})

end:
	return result, err
}

type SearchFilesOptions struct {
	Recursive   bool
	Pattern     string
	NamePattern string
	Extensions  []string
	FilesOnly   bool
	DirsOnly    bool
	MaxResults  int
}

func (t *SearchFilesTool) searchFiles(searchPath string, opts SearchFilesOptions) (results []FileSearchResult, err error) {
	var searchDir string

	if !t.IsAllowedPath(searchPath) {
		err = fmt.Errorf("access denied: path not allowed: %s", searchPath)
		goto end
	}

	searchDir, err = filepath.Abs(searchPath)
	if err != nil {
		goto end
	}

	err = filepath.Walk(searchDir, func(path string, info os.FileInfo, walkErr error) (err error) {
		var shouldInclude bool
		var result FileSearchResult

		if walkErr != nil {
			// Log but continue walking
			logger.Info("Walk error", "path", path, "error", walkErr)
			goto end
		}

		// Stop if we've hit the max results
		if 0 < opts.MaxResults && len(results) >= opts.MaxResults {
			err = filepath.SkipDir
			goto end
		}

		// Skip subdirectories if not recursive (but allow the root directory)
		if !opts.Recursive && info.IsDir() && path != searchDir {
			err = filepath.SkipDir
			goto end
		}

		// Apply file type filters
		if opts.FilesOnly && info.IsDir() {
			goto end
		}
		if opts.DirsOnly && !info.IsDir() {
			goto end
		}

		// Apply pattern matching
		shouldInclude = t.matchesFilters(info.Name(), opts)
		if !shouldInclude {
			goto end
		}

		result = FileSearchResult{
			Path:     path,
			Name:     info.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Format(time.RFC3339),
			IsDir:    info.IsDir(),
		}

		results = append(results, result)

	end:
		return err
	})

end:
	return results, err
}

func (t *SearchFilesTool) matchesFilters(fileName string, opts SearchFilesOptions) (matches bool) {
	// Pattern matching (case-insensitive substring)
	if opts.Pattern != "" {
		if !strings.Contains(strings.ToLower(fileName), strings.ToLower(opts.Pattern)) {
			goto end
		}
	}

	// Exact name pattern matching
	if opts.NamePattern != "" {
		matched, err := filepath.Match(opts.NamePattern, fileName)
		if err != nil || !matched {
			goto end
		}
	}

	// Extension filtering
	if len(opts.Extensions) > 0 {
		fileExt := strings.ToLower(filepath.Ext(fileName))
		extensionMatched := false

		for _, ext := range opts.Extensions {
			// Normalize extension (ensure it starts with .)
			normalizedExt := ext
			if !strings.HasPrefix(ext, ".") {
				normalizedExt = "." + ext
			}

			if strings.ToLower(normalizedExt) == fileExt {
				extensionMatched = true
				break
			}
		}

		if !extensionMatched {
			goto end
		}
	}

	matches = true

end:
	return matches
}
