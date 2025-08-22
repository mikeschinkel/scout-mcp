package mcputil

import (
	"context"
	_ "embed"
	"fmt"
	"reflect"
	"sync"
)

// payloadTypes contains a registry of payload types; we pick 3 as most we'll need
var payloadTypes = make(map[string]reflect.Type, 3)
var ptMutex = &sync.Mutex{} // Mutex to protect payload type registry

var _ Tool = (*StartSessionTool)(nil)

// RegisterPayloadType registers a payload type for start_session tool results.
func RegisterPayloadType(payload any) {
	t := reflect.TypeOf(payload)
	ptMutex.Lock()
	defer ptMutex.Unlock()
	payloadTypes[t.String()] = t
}

// PayloadCarrier interface for types that can carry payload information.
type PayloadCarrier interface {
	GetPayloadTypeName() string
	SetPayload(Payload)
}

// GetPayloadInfo extracts payload type information from a PayloadCarrier.
func GetPayloadInfo(g any) (t reflect.Type, tn string, _ PayloadCarrier) {
	carrier, ok := g.(PayloadCarrier)
	if ok {
		tn = carrier.GetPayloadTypeName()
		t = GetPayloadType(tn)
	}
	return t, tn, carrier
}

// GetPayloadType retrieves a registered payload type by name.
func GetPayloadType(pt string) reflect.Type {
	t, ok := payloadTypes[pt]
	if !ok {
		t = payloadTypes["*"+pt]
	}
	return t
}

// NewStartSessionTool creates a new start_session tool with the specified payload.
func NewStartSessionTool(p Payload) *StartSessionTool {
	return &StartSessionTool{
		Payload: p,
		ToolBase: NewToolBase(ToolOptions{
			Name:        "start_session",
			Description: "Start an MCP session and get comprehensive instructions for the MCP server effectively",
			Properties:  []Property{},
		}),
	}
}

type StartSessionTool struct {
	*ToolBase
	Payload Payload
}

var instructions = `🎯 MCP Session Started Successfully!

Your session token is valid for 24 hours and will be REQUIRED for all subsequent tool calls.

ON ANY SCOUT MCP TOOL ERROR
	1. You MUST IMMEDIATELY STOP and report the error. Provide the user with:
		A. The request you sent, in the JSONRPC 2.0 format a MCP Server expects,   
		B. The error message you received, and
		C. What you expected do happen.
	2. Log all tool failures in ./MCP_USABILITY_CONCERNS.md, then 
	3. Stop and wait for instructions from the user.

IMPORTANT INSTRUCTIONS:
1. **Session Token Required**: All tools (except start_session) require session_token parameter
2. **Token Expiration**: Tokens expire after 24 hours or when server restarts

`

// EnsurePreconditions bypasses session validation for start_session but runs other preconditions
func (t *StartSessionTool) EnsurePreconditions(context.Context, ToolRequest) (err error) {
	// start_session tool doesn't require any preconditions
	// Future non-session preconditions could be added here if needed
	return nil
}

func (t *StartSessionTool) Handle(_ context.Context, tr ToolRequest) (result ToolResult, err error) {
	var response StartSessionResult
	var ptn string

	logger.Info("Tool called", "tool", "start_session")

	// Create new session
	session := NewSession()
	err = session.Initialize()
	if err != nil {
		result = NewToolResultError(fmt.Errorf("failed to create session: %v", err))
		goto end
	}

	if t.Payload != nil {
		ptn = reflect.TypeOf(t.Payload).Elem().String()
		err = t.Payload.Initialize(t, tr)
	}
	if err != nil {
		result = NewToolResultError(err)
		goto end
	}

	// Build response
	response = StartSessionResult{
		SessionToken:    session.Token,
		TokenExpiresAt:  session.ExpiresAt,
		Instructions:    instructions,
		PayloadTypeName: ptn,
		Payload:         t.Payload,
		Message:         "MCP Session Started",
	}

	logger.Info("Tool completed",
		"tool", "start_session",
		"result", "success",
		"token_length", len(session.Token),
	)
	result = NewToolResultJSON(response)

end:
	return result, err
}

// generateQuickStartList creates a list of essential tools with their quick help descriptions
func (t *StartSessionTool) generateQuickStartList() []string {
	var tools []Tool
	var tool Tool
	var options ToolOptions
	var quickStartList []string

	tools = RegisteredTools()
	quickStartList = make([]string, 0)

	for _, tool = range tools {
		options = tool.Options()
		if options.QuickHelp != "" {
			quickStartList = append(quickStartList, fmt.Sprintf("%s - %s", options.Name, options.QuickHelp))
		}
	}

	return quickStartList
}

//
//// Default configurations
//
//func getPropertyDescription(prop Property) string {
//	// Extract description from property options
//	var options []PropertyOption
//	var opt PropertyOption
//
//	options = prop.PropertyOptions()
//	for _, opt = range options {
//		p, ok := opt.(DescriptionProperty)
//		if ok {
//			return p.Description
//		}
//	}
//	return "No description available"
//}
