package mcptools_test

import (
	"log/slog"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/testutil"
)

// TestEnvironment provides implementations of testutil.Environment for mcptools
type TestEnvironment struct{}

// SetLogger implements testutil.Environment
func (TestEnvironment) SetLogger(logger *slog.Logger) {
	mcptools.SetLogger(logger)
}

// Global instance for convenience
var testEnvironment = TestEnvironment{}

// setupTestEnv is a convenience function that wraps testutil.SetupUnitTest
func setupTestEnv(t *testing.T) (tempDir string, cleanup func()) {
	t.Helper()
	return testutil.SetupUnitTest(t, testEnvironment)
}
