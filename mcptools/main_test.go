package mcptools_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/langutil/golang"
	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
)

var logger *slog.Logger

func TestMain(m *testing.M) {
	// Setup code here if needed
	// For example: initialize test data, mock services, etc.
	logger := testutil.NewTestLogger()
	mcptools.SetLogger(logger)
	mcputil.SetLogger(logger)
	golang.SetLogger(logger)

	// Run tests
	code := m.Run()

	// Cleanup code here if needed

	os.Exit(code)
}
