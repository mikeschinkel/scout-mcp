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
	var operation, riskLevel, impactSummary, previewContent string
	var files []string

	logger.Info("Tool called", "tool", "request_approval")

	operation, err = mcputil.String("operation", "").String(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	files, err = FilesProperty.StringSlice(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	riskLevel, _ = mcputil.String("risk_level", "").String(req)
	impactSummary, _ = mcputil.String("impact_summary", "").String(req)
	previewContent, _ = mcputil.String("preview_content", "").String(req)

	if riskLevel == "" {
		riskLevel = "medium"
	}

	logger.Info("Tool completed", "tool", "request_approval", "operation", operation, "risk_level", riskLevel)
	result = mcputil.NewToolResultJSON(map[string]any{
		"status":          "approval_requested",
		"operation":       operation,
		"risk_level":      riskLevel,
		"files_affected":  len(files),
		"files":           files,
		"impact_summary":  impactSummary,
		"preview_content": previewContent,
		"message":         "Approval request logged for manual review",
		"note":            "This is a stub implementation. In a real system, this would display rich formatted approval UI and wait for user confirmation.",
	})

end:
	return result, err
}
