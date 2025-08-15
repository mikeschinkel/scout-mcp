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
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "delete_files",
			Description: "Delete file or directory from allowed directories",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				PathProperty.Required(),
				RecursiveProperty,
			},
		}),
	})
}

// DeleteFileTool deletes files or directories from allowed directories with optional recursive deletion.
type DeleteFileTool struct {
	*mcputil.ToolBase
}

// Handle processes the delete_files tool request and removes the specified file or directory.
func (t *DeleteFileTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var recursive bool
	var fileInfo os.FileInfo
	var fileType string

	logger.Info("Tool called", "tool", "delete_files")

	filePath, err = PathProperty.String(req)
	if err != nil {
		goto end
	}

	recursive, err = RecursiveProperty.Bool(req)
	if err != nil {
		goto end
	}

	logger.Info("Tool arguments parsed", "tool", "delete_files", "path", filePath, "recursive", recursive)

	// Check path is allowed
	if !t.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	// Check if file/directory exists
	fileInfo, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		err = fmt.Errorf("file or directory does not exist: %s", filePath)
		goto end
	}
	if err != nil {
		err = fmt.Errorf("error checking file: %v", err)
		goto end
	}

	// Determine what we're deleting
	if fileInfo.IsDir() {
		fileType = "directory"
		if !recursive {
			err = fmt.Errorf("cannot delete directory without recursive flag: %s", filePath)
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
		err = fmt.Errorf("failed to delete %s: %v", fileType, err)
		goto end
	}

	logger.Info("Tool completed", "tool", "delete_files", "success", true, "path", filePath, "type", fileType)
	result = mcputil.NewToolResultJSON(map[string]any{
		"success":      true,
		"deleted_path": filePath,
		"file_type":    fileType,
		"message":      fmt.Sprintf("%s deleted successfully: %s", titleCase(fileType), filePath),
	})
end:
	return result, err
}
