package mcputil

import (
	"context"
)

// CallTool is a helper to call a tool for unit tests (bypasses session validation and other preconditions).
// This function directly invokes the tool's Handle method without going through the normal
// framework-level precondition checks, making it suitable for isolated unit testing of tool logic.
func CallTool(tool Tool, req ToolRequest) (ToolResult, error) {
	// For unit tests, we bypass EnsurePreconditions to focus on business logic
	// and avoid framework concerns like session validation
	return tool.Handle(context.Background(), req)
}
