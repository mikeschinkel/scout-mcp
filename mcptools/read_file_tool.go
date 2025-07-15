package mcptools

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*ReadFileTool)(nil)

func init() {
	mcputil.RegisterTool(&ReadFileTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "read_file",
			Description: "Read contents of a file from an allowed directory",
			Properties: []mcputil.Property{
				mcputil.String("path", "File path to read").Required(),
			},
		}),
	})
}

type ReadFileTool struct {
	*toolBase
}

func (t *ReadFileTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var content string
	var allowed bool

	filePath, err = req.RequireString("path")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	// Check path is allowed
	allowed, err = t.IsAllowedPath(filePath)
	if err != nil {
		goto end
	}

	if !allowed {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	content, err = readFile(t.Config(), filePath)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	result = mcputil.NewToolResultText(content)

end:
	return result, err
}
