package mcptools

import (
	"context"
	"fmt"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*GenerateApprovalTokenTool)(nil)

func init() {
	mcputil.RegisterTool(&GenerateApprovalTokenTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "generate_approval_token",
			Description: "Generate approval token after user confirmation",
			Properties: []mcputil.Property{
				mcputil.Array("file_actions", "File actions approved").Required(),
				mcputil.Array("operations", "Operations approved (create, update, delete)").Required(),
				mcputil.String("session_id", "Session identifier for this approval"),
			},
		}),
	})
}

type GenerateApprovalTokenTool struct {
	*toolBase
}

func (t *GenerateApprovalTokenTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var token string
	var fileActions []FileAction
	var operations []string
	var sessionID string

	logger.Info("Tool called", "tool", "generate_approval_token")

	fileActions, err = getFileActions(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	operations, err = getOperations(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}
	sessionID = req.GetString("session_id", "")

	// Generate the JWT token
	token, err = generateApprovalToken(TokenRequest{
		FileActions: fileActions,
		Operations:  operations,
		SessionID:   sessionID,
		ExpiresIn:   time.Hour,
	})
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to generate token: %v", err))
		goto end
	}

	result = mcputil.NewToolResultText(fmt.Sprintf("✅ Approval token generated (expires in 1 hour)\n🔑 Token: %s", token))

end:
	return result, err
}

func generateApprovalToken(_ TokenRequest) (token string, err error) {
	// JWT token generation logic
	return token, err
}
