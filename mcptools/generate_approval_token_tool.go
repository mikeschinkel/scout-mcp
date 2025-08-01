package mcptools

import (
	"context"
	"fmt"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*GenerateApprovalTokenTool)(nil)

var (
	fileActionsProperty = mcputil.Array("file_actions", "File actions approved")
	operationsProperty  = mcputil.Array("operations", "Operations approved (create, update, delete)")
)

func init() {
	mcputil.RegisterTool(&GenerateApprovalTokenTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "generate_approval_token",
			Description: "Generate approval token after user confirmation",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				fileActionsProperty.Required(),
				operationsProperty.Required(),
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

	fileActions, err = mcputil.TypedPropertySlice[FileAction](req, fileActionsProperty)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	operations, err = operationsProperty.StringSlice(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}
	sessionID, _ = SessionIdProperty.String(req)

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
	if err != nil {
		// TODO Verify there are no use-cases where this should generate a real error
		err = nil
	}
	return result, err
}

func generateApprovalToken(_ TokenRequest) (token string, err error) {
	// JWT token generation logic
	return token, err
}
