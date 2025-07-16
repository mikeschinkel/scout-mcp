package mcptools

import (
	"context"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*RequestApprovalTool)(nil)

func init() {
	mcputil.RegisterTool(&RequestApprovalTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "request_approval",
			Description: "Request user approval with rich visual formatting",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				mcputil.String("operation", "Brief operation description").Required(),
				FilesProperty.Required(),
				mcputil.String("preview_content", "Code preview or diff content"),
				mcputil.String("risk_level", "Risk level: low, medium, or high"),
				mcputil.String("impact_summary", "Summary of what will change"),
			},
		}),
	})
}

type RequestApprovalTool struct {
	*toolBase
}

func (t *RequestApprovalTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var example string
	result = mcputil.NewToolResultJSON(example)
	//end:
	return result, err
}
