package mcptools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*CreateFileTool)(nil)

func init() {
	mcputil.RegisterTool(&CreateFileTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "create_file",
			Description: "Create a new file in allowed directories",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				FilepathProperty.Required(),
				NewContentProperty.Required(),
				mcputil.Bool("create_dirs", "Create parent directories if needed"),
			},
		}),
	})
}

type CreateFileTool struct {
	*toolBase
}

func (t *CreateFileTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var content string
	var createDirs bool
	var allowed bool
	var fileDir string

	logger.Info("Tool called", "tool", "create_file")

	filePath, err = req.RequireString("filepath")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	content, err = req.RequireString("new_content")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	createDirs = req.GetBool("create_dirs", false)

	logger.Info("Tool arguments parsed", "tool", "create_file", "path", filePath, "create_dirs", createDirs, "content_length", len(content))

	// Check path is allowed
	allowed, err = t.IsAllowedPath(filePath)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("path validation failed: %v", err))
		goto end
	}
	if !allowed {
		result = mcputil.NewToolResultError(fmt.Errorf("access denied: path not allowed: %s", filePath))
		goto end
	}

	// Check if file already exists
	_, err = os.Stat(filePath)
	if err == nil {
		result = mcputil.NewToolResultError(fmt.Errorf("file already exists: %s", filePath))
		goto end
	}
	if !os.IsNotExist(err) {
		result = mcputil.NewToolResultError(fmt.Errorf("error checking file: %v", err))
		goto end
	}

	// Create parent directories if requested
	if createDirs {
		fileDir = filepath.Dir(filePath)
		err = os.MkdirAll(fileDir, 0755)
	}
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to create directories: %v", err))
		goto end
	}

	// Create the file
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to create file: %v", err))
		goto end
	}

	logger.Info("Tool completed", "tool", "create_file", "success", true, "path", filePath)
	result = mcputil.NewToolResultText(fmt.Sprintf("File created successfully: %s (%d bytes)", filePath, len(content)))
end:
	return result, err
}
