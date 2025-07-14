package mcptools

import (
	"context"
	"fmt"
	"os"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*DeleteFileTool)(nil)

func init() {
	mcputil.RegisterTool(&DeleteFileTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "delete_file",
			Description: "Delete file or directory from allowed directories",
			Properties: []mcputil.Property{
				mcputil.String("path", "File or directory path to delete").Required(),
				mcputil.Bool("recursive", "Delete directory recursively"),
			},
		}),
	})
}

type DeleteFileTool struct {
	*toolBase
}

func (t *DeleteFileTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var recursive bool
	var allowed bool
	var fileInfo os.FileInfo
	var fileType string

	logger.Info("Tool called", "tool", "delete_file")

	filePath, err = req.RequireString("path")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	recursive = req.GetBool("recursive", false)

	logger.Info("Tool arguments parsed", "tool", "delete_file", "path", filePath, "recursive", recursive)

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

	// Check if file/directory exists
	fileInfo, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		result = mcputil.NewToolResultError(fmt.Errorf("file or directory does not exist: %s", filePath))
		goto end
	}
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("error checking file: %v", err))
		goto end
	}

	// Determine what we're deleting
	if fileInfo.IsDir() {
		fileType = "directory"
		if !recursive {
			result = mcputil.NewToolResultError(fmt.Errorf("cannot delete directory without recursive flag: %s", filePath))
			goto end
		}
		// Use RemoveAll for recursive directory deletion
		err = os.RemoveAll(filePath)
	} else {
		fileType = "file"
		// Use Remove for single file deletion
		err = os.Remove(filePath)
	}

	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to delete %s: %v", fileType, err))
		goto end
	}

	logger.Info("Tool completed", "tool", "delete_file", "success", true, "path", filePath, "type", fileType)
	result = mcputil.NewToolResultText(fmt.Sprintf("%s deleted successfully: %s",
		titleCase(fileType),
		filePath,
	))
end:
	return result, err
}
