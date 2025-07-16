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

func init() {
	mcputil.RegisterTool(&SearchFilesTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "search_files",
			Description: "Search for files and directories in allowed paths with filtering options",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				PathProperty.Required(),
				RecursiveProperty,
				mcputil.Array("extensions", "Filter by file extensions (e.g., ['.go', '.txt'])"),
				mcputil.String("pattern", "Name pattern to match (case-insensitive substring)"),
				mcputil.String("name_pattern", "Exact filename pattern to match"),
				mcputil.Bool("files_only", "Return only files, not directories"),
				mcputil.Bool("dirs_only", "Return only directories, not files"),
				mcputil.Number("max_results", "Maximum number of results to return"),
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
	var allowed bool
	var results []FileSearchResult

	logger.Info("Tool called", "tool", "search_files")

	searchPath, err = req.RequireString("path")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	recursive = req.GetBool("recursive", false)
	pattern = req.GetString("pattern", "")
	namePattern = req.GetString("name_pattern", "")
	filesOnly = req.GetBool("files_only", false)
	dirsOnly = req.GetBool("dirs_only", false)
	maxResults = req.GetInt("max_results", 1000)

	// Get extensions array if provided
	extensions, err = getStringSlice(req, "extensions")
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
	allowed, err = t.IsAllowedPath(searchPath)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("path validation failed: %v", err))
		goto end
	}

	if !allowed {
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
	var allowed bool

	allowed, err = t.IsAllowedPath(searchPath)
	if err != nil {
		goto end
	}

	if !allowed {
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
		if len(results) >= opts.MaxResults {
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
		shouldInclude = t.matchesFilters(info.Name(), path, opts)
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

func (t *SearchFilesTool) matchesFilters(fileName, fullPath string, opts SearchFilesOptions) (matches bool) {
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
