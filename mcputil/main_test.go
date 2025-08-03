package mcputil_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
)

func TestMain(m *testing.M) {
	// Setup code here if needed
	// For example: initialize test data, mock services, etc.

	mcputil.SetLogger(testutil.NewTestLogger())

	// Need to register for standalone testing
	// When paired with a app-specific tools package this will be registered.
	mcputil.RegisterTool(mcputil.NewStartSessionTool(nil))

	// Run tests
	code := m.Run()

	// Cleanup code here if needed

	os.Exit(code)
}
