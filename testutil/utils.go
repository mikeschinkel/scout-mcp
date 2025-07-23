package testutil

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// CallTool is a helper to call a tool for unit tests (bypasses session validation and other preconditions)
func CallTool(tool mcputil.Tool, req mcputil.ToolRequest) (mcputil.ToolResult, error) {
	// For unit tests, we bypass EnsurePreconditions to focus on business logic
	// and avoid framework concerns like session validation
	return tool.Handle(context.Background(), req)
}

// MaybeRemove wraps calls to os.Remove and logs errors that are other than not-exists
func MaybeRemove(t *testing.T, fp string) {
	var ok bool

	t.Helper()
	err := os.Remove(fp)
	if err == nil {
		goto end
	}
	_, ok = err.(*fs.PathError)
	if ok {
		goto end
	}

	t.Error(err)
end:
}
