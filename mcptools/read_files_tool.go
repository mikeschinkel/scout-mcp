package mcptools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*ReadFilesTool)(nil)

func init() {
	mcputil.RegisterTool(&ReadFilesTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "read_files",
			Description: "Read contents of multiple files and/or directories with filtering options",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				PathsProperty.Required(),
				mcputil.Array("extensions", "Filter by file extensions (e.g., ['.go', '.txt']) - applies to directories only"),
				RecursiveProperty,
				mcputil.String("pattern", "Filename pattern to match (case-insensitive substring) - applies to directories only"),
				mcputil.Number("max_files", "Maximum number of files to read (default: 100)"),
			},
		}),
	})
}

type ReadFilesTool struct {
	*toolBase
}

func (t *ReadFilesTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var paths []string
	var extensions []string
	var recursive bool
	var pattern string
	var maxFiles int
	var fileResults []FileReadResult
	var totalSize int64
	var errors []string

	logger.Info("Tool called", "tool", "read_files")

	// Get paths array
	paths, err = getStringSlice(req, "paths")
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("invalid paths array: %v", err))
		goto end
	}

	if len(paths) == 0 {
		result = mcputil.NewToolResultError(fmt.Errorf("paths array cannot be empty"))
		goto end
	}

	// Get optional parameters
	extensions, err = getStringSlice(req, "extensions")
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("invalid extensions array: %v", err))
		goto end
	}

	recursive = req.GetBool("recursive", false)
	pattern = req.GetString("pattern", "")
	maxFiles = req.GetInt("max_files", 100)

	logger.Info("Tool arguments parsed",
		"tool", "read_files",
		"paths", paths,
		"extensions", extensions,
		"recursive", recursive,
		"pattern", pattern,
		"max_files", maxFiles)

	fileResults, totalSize, errors, err = t.readMultiplePaths(paths, ReadFilesOptions{
		Extensions: extensions,
		Recursive:  recursive,
		Pattern:    pattern,
		MaxFiles:   maxFiles,
	})
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	logger.Info("Tool completed", "tool", "read_files", "files_read", len(fileResults), "total_size", totalSize)

	result = mcputil.NewToolResultJSON(map[string]any{
		"files":       fileResults,
		"total_files": len(fileResults),
		"total_size":  totalSize,
		"errors":      errors,
		"paths":       paths,
		"extensions":  extensions,
		"recursive":   recursive,
		"pattern":     pattern,
		"max_files":   maxFiles,
		"truncated":   len(fileResults) >= maxFiles,
	})

end:
	return result, err
}

type ReadFilesOptions struct {
	Extensions []string
	Recursive  bool
	Pattern    string
	MaxFiles   int
}

type FileReadResult struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Size    int64  `json:"size"`
	Error   string `json:"error,omitempty"`
}

func (t *ReadFilesTool) readMultiplePaths(paths []string, opts ReadFilesOptions) (results []FileReadResult, totalSize int64, errors []string, err error) {
	var filesToRead []string
	var path string
	var allowed bool

	// First pass: collect all files to read
	for _, path = range paths {
		// Check if path is allowed
		allowed, err = t.IsAllowedPath(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("path validation failed for %s: %v", path, err))
			continue
		}

		if !allowed {
			errors = append(errors, fmt.Sprintf("access denied: path not allowed: %s", path))
			continue
		}

		// Check if it's a file or directory
		info, statErr := os.Stat(path)
		if statErr != nil {
			errors = append(errors, fmt.Sprintf("cannot access %s: %v", path, statErr))
			continue
		}

		if info.IsDir() {
			// Directory - find files within it
			dirFiles, dirErr := t.findFilesInDirectory(path, opts)
			if dirErr != nil {
				errors = append(errors, fmt.Sprintf("error reading directory %s: %v", path, dirErr))
				continue
			}
			filesToRead = append(filesToRead, dirFiles...)
		} else {
			// Single file
			filesToRead = append(filesToRead, path)
		}

		// Stop if we've reached the limit
		if len(filesToRead) >= opts.MaxFiles {
			filesToRead = filesToRead[:opts.MaxFiles]
			break
		}
	}

	// Second pass: read all collected files
	results = make([]FileReadResult, 0, len(filesToRead))
	for _, filePath := range filesToRead {
		var content string
		var fileInfo os.FileInfo
		var readErr error

		fileInfo, readErr = os.Stat(filePath)
		if readErr != nil {
			results = append(results, FileReadResult{
				Path:  filePath,
				Name:  filepath.Base(filePath),
				Error: fmt.Sprintf("cannot stat file: %v", readErr),
			})
			continue
		}

		content, readErr = t.readFile(filePath)
		if readErr != nil {
			results = append(results, FileReadResult{
				Path:  filePath,
				Name:  filepath.Base(filePath),
				Size:  fileInfo.Size(),
				Error: fmt.Sprintf("cannot read file: %v", readErr),
			})
			continue
		}

		results = append(results, FileReadResult{
			Path:    filePath,
			Name:    filepath.Base(filePath),
			Content: content,
			Size:    fileInfo.Size(),
		})

		totalSize += fileInfo.Size()
	}

	return results, totalSize, errors, err
}

func (t *ReadFilesTool) findFilesInDirectory(dirPath string, opts ReadFilesOptions) (files []string, err error) {
	var entries []os.DirEntry

	entries, err = os.ReadDir(dirPath)
	if err != nil {
		goto end
	}

	for _, entry := range entries {
		var fullPath string
		var shouldInclude bool

		fullPath = filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// Handle subdirectories if recursive
			if opts.Recursive {
				var subFiles []string
				subFiles, err = t.findFilesInDirectory(fullPath, opts)
				if err != nil {
					// Log error but continue with other directories
					logger.Info("Error reading subdirectory", "path", fullPath, "error", err)
					continue
				}
				files = append(files, subFiles...)
			}
			continue
		}

		// Apply filters to files
		shouldInclude = t.matchesFileFilters(entry.Name(), opts)
		if shouldInclude {
			files = append(files, fullPath)
		}
	}

end:
	return files, err
}

func (t *ReadFilesTool) matchesFileFilters(fileName string, opts ReadFilesOptions) (matches bool) {
	// Pattern matching (case-insensitive substring)
	if opts.Pattern != "" {
		if !strings.Contains(strings.ToLower(fileName), strings.ToLower(opts.Pattern)) {
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
