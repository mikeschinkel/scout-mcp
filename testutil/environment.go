package testutil

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Environment represents the interface for test environment setup
type Environment interface {
	SetLogger(logger *slog.Logger)
}

// SetupUnitTest creates a test environment for unit testing tool business logic
// Unit tests bypass session validation and other framework concerns
func SetupUnitTest(t *testing.T, env Environment) (tempDir string, cleanup func()) {
	t.Helper()

	// Use quiet logger that discards all output
	env.SetLogger(QuietLogger())

	// Create temp directory
	var err error
	tempDir, err = os.MkdirTemp("", "scout-mcp-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Create some test files
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err, "Failed to create test file")

	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err, "Failed to create subdirectory")

	subFile := filepath.Join(subDir, "test.go")
	err = os.WriteFile(subFile, []byte("package main\n\nfunc main() {}\n"), 0644)
	require.NoError(t, err, "Failed to create Go test file")

	cleanup = func() {
		MaybeRemove(t, tempDir)
	}

	return tempDir, cleanup
}
