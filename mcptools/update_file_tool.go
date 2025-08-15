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
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "update_file",
			Description: "Update existing file in allowed directories",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				FilepathProperty.Required(),
				NewContentProperty.Required(),
			},
		}),
	})
}

type UpdateFileTool struct {
	*mcputil.ToolBase
}

func (t *UpdateFileTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var content string
	var fileInfo os.FileInfo
	var oldSize int64

	logger.Info("Tool called", "tool", "update_file")

	filePath, err = FilepathProperty.String(req)
	if err != nil {
		goto end
	}

	content, err = NewContentProperty.String(req)
	if err != nil {
		goto end
	}

	logger.Info("Tool arguments parsed", "tool", "update_file", "path", filePath, "content_length", len(content))

	// Check path is allowed
	if !t.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	// Check if file exists
	fileInfo, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		err = fmt.Errorf("file does not exist: %s", filePath)
		goto end
	}
	if err != nil {
		err = fmt.Errorf("error checking file: %v", err)
		goto end
	}

	// Don't allow updating directories
	if fileInfo.IsDir() {
		err = fmt.Errorf("cannot update directory: %s", filePath)
		goto end
	}

	oldSize = fileInfo.Size()

	// Update the file
	err = os.WriteFile(filePath, []byte(content), fileInfo.Mode())
	if err != nil {
		err = fmt.Errorf("failed to update file: %v", err)
		goto end
	}

	logger.Info("Tool completed", "tool", "update_file", "success", true, "path", filePath)
	result = mcputil.NewToolResultJSON(map[string]any{
		"success":   true,
		"file_path": filePath,
		"old_size":  oldSize,
		"new_size":  len(content),
		"message":   fmt.Sprintf("File updated successfully: %s (%d -> %d bytes)", filePath, oldSize, len(content)),
	})
end:
	return result, err
}
