package mcptools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*CreateFileTool)(nil)

func init() {
	mcputil.RegisterTool(&CreateFileTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "create_file",
			Description: "Create a new file in allowed directories",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				RequiredPathProperty,
				RequiredNewContentProperty,
				CreateDirsProperty,
			},
		}),
	})
}

// CreateFileTool creates new files in allowed directories with optional parent directory creation.
type CreateFileTool struct {
	*mcputil.ToolBase
}

// Handle processes the create_file tool request and creates a new file with the specified content.
func (t *CreateFileTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var content string
	var createDirs bool
	var fileDir string

	logger.Info("Tool called", "tool", "create_file")

	filePath, err = FilepathProperty.String(req)
	if err != nil {
		goto end
	}

	content, err = NewContentProperty.String(req)
	if err != nil {
		goto end
	}

	createDirs, _ = CreateDirsProperty.Bool(req)

	logger.Info("Tool arguments parsed", "tool", "create_file", "path", filePath, "create_dirs", createDirs, "content_length", len(content))

	// Check path is allowed
	if !t.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	// Check if file already exists
	err = checkFileExists(filePath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// This is what we want
		err = nil
	case errors.Is(err, os.ErrExist):
		err = fmt.Errorf("file already exists: %s", filePath)
		goto end
	default:
		err = fmt.Errorf("error checking file: %v", err)
		goto end
	}

	// Create parent directories if requested
	if createDirs {
		fileDir = filepath.Dir(filePath)
		err = os.MkdirAll(fileDir, 0755)
	}
	if err != nil {
		err = fmt.Errorf("failed to create directories: %v", err)
		goto end
	}

	// Create the file
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		err = fmt.Errorf("failed to create file: %v", err)
		goto end
	}

	logger.Info("Tool completed", "tool", "create_file", "success", true, "path", filePath)
	result = mcputil.NewToolResultJSON(map[string]any{
		"success":   true,
		"file_path": filePath,
		"size":      len(content),
		"message":   fmt.Sprintf("File created successfully: %s (%d bytes)", filePath, len(content)),
	})
end:
	return result, err
}
