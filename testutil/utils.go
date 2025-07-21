package testutil

import (
	"context"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// CallTool is a helper to call a tool for unit tests (bypasses session validation and other preconditions)
func CallTool(tool mcputil.Tool, req mcputil.ToolRequest) (mcputil.ToolResult, error) {
	// For unit tests, we bypass EnsurePreconditions to focus on business logic
	// and avoid framework concerns like session validation
	return tool.Handle(context.Background(), req)
}

// Must wraps calls that return an error and only an error and throws logs the error if it occurs
func Must(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
