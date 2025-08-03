package mcptools

import (
	"context"
	"fmt"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*GenerateApprovalTokenTool)(nil)

var (
	fileActionsProperty = mcputil.Array("file_actions", "File actions approved", mcputil.DefaultArray[FileAction]{})
	operationsProperty  = mcputil.Array("operations", "Operations approved (create, update, delete)", mcputil.DefaultArray[string]{})
)

type FileAction struct {
	Action  string `json:"action"`  // create, update, delete
	Path    string `json:"path"`    // file path
	Purpose string `json:"purpose"` // why this file is being modified
}

func (fa FileAction) Icon() string {
	switch fa.Action {
	case "create":
		return "✨"
	case "update", "modify":
		return "📝"
	case "delete":
		return "🗑️"
	case "move", "rename":
		return "📦"
	default:
		return "📄"
	}
}

type TokenRequest struct {
	FileActions []FileAction
	Operations  []string
	SessionID   string
	ExpiresIn   time.Duration
}

func init() {
	mcputil.RegisterTool(&GenerateApprovalTokenTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "generate_approval_token",
			Description: "Generate approval token after user confirmation",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				fileActionsProperty,
				operationsProperty,
			},
		}),
	})
}

type GenerateApprovalTokenTool struct {
	*mcputil.ToolBase
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

	result = mcputil.NewToolResultJSON(map[string]any{
		"success":    true,
		"token":      token,
		"expires_in": "1 hour",
		"message":    "Approval token generated successfully",
	})

end:
	if err != nil {
		// TODO Verify there are no use-cases where this should generate a real error
		err = nil
	}
	return result, err
}

func generateApprovalToken(req TokenRequest) (token string, err error) {
	// For now, generate a simple mock token
	// In a real implementation, this would generate a proper JWT token
	token = fmt.Sprintf("mock-approval-token-%d-ops-%d-files-%s",
		len(req.Operations),
		len(req.FileActions),
		req.SessionID[:min(8, len(req.SessionID))])
	return token, err
}
