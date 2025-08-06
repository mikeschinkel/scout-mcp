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
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "read_files",
			Description: "Read contents of multiple files and/or directories with filtering options",
			QuickHelp:   "Read multiple files efficiently (read before updating)",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				PathsProperty.Required(),
				ExtensionsProperty.Description("Filter by file extensions (e.g., ['.go', '.txt']) - applies to directories only"),
				RecursiveProperty,
				PatternProperty.Description("Filename pattern to match (case-insensitive substring) - applies to directories only"),
				MaxFilesProperty,
			},
		}),
	})
}

type ReadFilesTool struct {
	*mcputil.ToolBase
}

func (t *ReadFilesTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var paths []string
	var extensions []string
	var recursive bool
	var pattern string
	var maxFiles int
	var fileResults []FileReadResult
	var totalSize int64
	var errs []error

	logger.Info("Tool called", "tool", "read_files")

	paths, err = PathsProperty.StringSlice(req)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("invalid paths array: %v", err))
		goto end
	}

	if len(paths) == 0 {
		result = mcputil.NewToolResultError(fmt.Errorf("paths array cannot be empty"))
		goto end
	}

	extensions, err = ExtensionsProperty.StringSlice(req)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("invalid extensions array: %v", err))
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

	maxFiles, err = MaxFilesProperty.Int(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	logger.Info("Tool arguments parsed",
		"tool", "read_files",
		"paths", paths,
		"extensions", extensions,
		"recursive", recursive,
		"pattern", pattern,
		"max_files", maxFiles)

	fileResults, totalSize, errs, err = t.readMultiplePaths(paths, ReadFilesOptions{
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
		"paths":       paths,
		"extensions":  extensions,
		"recursive":   recursive,
		"pattern":     pattern,
		"max_files":   maxFiles,
		"truncated":   len(fileResults) >= maxFiles,
		"errors":      errorsStringSlice(errs),
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

func (t *ReadFilesTool) readPath(path string, opts ReadFilesOptions) (entries []string, err error) {
	var info os.FileInfo

	// Check if path is allowed
	if !t.IsAllowedPath(path) {
		err = fmt.Errorf("access denied: path not allowed: %s", path)
		goto end
	}

	// Check if it's a file or directory
	info, err = os.Stat(path)
	if err != nil {
		err = fmt.Errorf("cannot access %s: %v", path, err)
		goto end
	}

	if !info.IsDir() {
		// Single file
		entries = append(entries, path)
		goto end
	}

	// Directory - find files within it
	entries, err = t.findFilesInDirectory(path, opts)
	if err != nil {
		err = fmt.Errorf("error reading directory %s: %v", path, err)
		goto end
	}
	entries = append(entries, entries...)

end:
	return entries, err
}

func (t *ReadFilesTool) readMultiplePaths(paths []string, opts ReadFilesOptions) (results []FileReadResult, totalSize int64, errs []error, err error) {
	var filesToRead, entries []string
	var path string

	// First pass: collect all files to read
	for _, path = range paths {
		entries, err = t.readPath(path, opts)
		if err != nil {
			errs = append(errs, err)
			err = nil
			continue
		}
		filesToRead = append(filesToRead, entries...)
		if len(filesToRead) >= opts.MaxFiles {
			// Stop if we've reached the limit
			filesToRead = filesToRead[:opts.MaxFiles]
			break
		}
	}

	// Second pass: read all collected files
	results = make([]FileReadResult, 0, len(filesToRead))
	for _, filePath := range filesToRead {
		var content []byte
		var fileInfo os.FileInfo
		var err error

		fileInfo, err = os.Stat(filePath)
		if err != nil {
			results = append(results, FileReadResult{
				Path:  filePath,
				Name:  filepath.Base(filePath),
				Error: fmt.Sprintf("cannot stat file: %v", err),
			})
			err = nil
			continue
		}

		content, err = os.ReadFile(filePath)
		if err != nil {
			results = append(results, FileReadResult{
				Path:  filePath,
				Name:  filepath.Base(filePath),
				Size:  fileInfo.Size(),
				Error: fmt.Sprintf("cannot read file: %v", err),
			})
			err = nil
			continue
		}

		results = append(results, FileReadResult{
			Path:    filePath,
			Name:    filepath.Base(filePath),
			Content: string(content),
			Size:    fileInfo.Size(),
		})

		totalSize += fileInfo.Size()

	}
	return results, totalSize, errs, err
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
