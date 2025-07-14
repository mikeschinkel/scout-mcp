package mcptools

import (
	"context"
	"fmt"
	"os"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*UpdateFileTool)(nil)

func init() {
	mcputil.RegisterTool(&UpdateFileTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "update_file",
			Description: "Update existing file in allowed directories",
			Properties: []mcputil.Property{
				mcputil.String("path", "File path to update").Required(),
				mcputil.String("content", "New file content").Required(),
			},
		}),
	})
}

type UpdateFileTool struct {
	*toolBase
}

func (t *UpdateFileTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var content string
	var allowed bool
	var fileInfo os.FileInfo
	var oldSize int64

	logger.Info("Tool called", "tool", "update_file")

	filePath, err = req.RequireString("path")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	content, err = req.RequireString("content")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	logger.Info("Tool arguments parsed", "tool", "update_file", "path", filePath, "content_length", len(content))

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

	// Check if file exists
	fileInfo, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		result = mcputil.NewToolResultError(fmt.Errorf("file does not exist: %s", filePath))
		goto end
	}
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("error checking file: %v", err))
		goto end
	}

	// Don't allow updating directories
	if fileInfo.IsDir() {
		result = mcputil.NewToolResultError(fmt.Errorf("cannot update directory: %s", filePath))
		goto end
	}

	oldSize = fileInfo.Size()

	// Update the file
	err = os.WriteFile(filePath, []byte(content), fileInfo.Mode())
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to update file: %v", err))
		goto end
	}

	logger.Info("Tool completed", "tool", "update_file", "success", true, "path", filePath)
	result = mcputil.NewToolResultText(fmt.Sprintf("File updated successfully: %s (%d -> %d bytes)", filePath, oldSize, len(content)))
end:
	return result, err
}
