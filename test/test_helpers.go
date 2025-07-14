package test

import (
	"os"
	"testing"
)

// getServerPath returns the path to the scout-mcp binary
func getServerPath(t *testing.T) string {
	// Try the compiled binary first
	if _, err := os.Stat(ServerBinaryPath); err == nil {
		return ServerBinaryPath
	}

	// Try fallback path for running from subdirectories
	if _, err := os.Stat(ServerBinaryPathFallback); err == nil {
		return ServerBinaryPathFallback
	}

	// Check if source exists
	if _, err := os.Stat(ServerSourcePath); err == nil {
		t.Log(BuildFromSourceMsg)
		t.Skip(BinaryNotFoundMsg)
	}

	t.Fatal(SourceNotFoundMsg)
	return ""
}
