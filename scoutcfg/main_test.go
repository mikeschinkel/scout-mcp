package scoutcfg_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/scoutcfg"
	"github.com/mikeschinkel/scout-mcp/testutil"
)

func TestMain(m *testing.M) {
	// Setup code here if needed
	// For example: initialize test data, mock services, etc.

	scoutcfg.SetLogger(testutil.NewTestLogger())

	// Run tests
	code := m.Run()

	// Cleanup code here if needed

	os.Exit(code)
}
