package test

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getExpectedToolsFromFilesystem discovers tool names by scanning for *_tool.go files
// in mcptools/ and mcputil/ directories. This ensures tests catch when tool files
// exist but fail to register due to syntax errors or other issues.
// Tool names are deduplicated since the same tool may exist in multiple directories.
func getExpectedToolsFromFilesystem(t *testing.T) []string {
	// Tool directories to scan
	toolDirs := []string{
		"../mcptools",
		"../mcputil",
	}

	// Use a map to deduplicate tool names
	toolMap := make(map[string]struct{})

	for _, dir := range toolDirs {
		pattern := filepath.Join(dir, "*_tool.go")
		matches, err := filepath.Glob(pattern)
		require.NoError(t, err, "Failed to glob tool files in %s", dir)

		for _, match := range matches {
			// Extract tool name from filename
			// e.g., "analyze_files_tool.go" -> "analyze_files"
			filename := filepath.Base(match)
			toolName := strings.TrimSuffix(filename, "_tool.go")
			toolMap[toolName] = struct{}{}
		}
	}

	// Convert map keys to sorted slice
	expectedTools := make([]string, 0, len(toolMap))
	for toolName := range toolMap {
		expectedTools = append(expectedTools, toolName)
	}
	sort.Strings(expectedTools)
	return expectedTools
}

// TestToolRegistrationCompleteness verifies that all *_tool.go files are properly registered
// This test catches syntax errors or other issues that prevent tools from being registered
func TestToolRegistrationCompleteness(t *testing.T) {
	// Get expected tools from filesystem
	expected := getExpectedToolsFromFilesystem(t)

	// Get actual registered tool names
	actualTools := mcputil.GetRegisteredToolNames()
	sort.Strings(actualTools)

	// Verify counts match
	assert.Equal(t, len(expected), len(actualTools), "Number of expected tools (%d) should match actual registered tools (%d)", len(expected), len(actualTools))

	// Verify all expected tools are registered
	for _, expectedTool := range expected {
		assert.Contains(t, actualTools, expectedTool, "Expected tool %s should be registered", expectedTool)
	}

	// Verify no unexpected tools are registered
	for _, actualTool := range actualTools {
		assert.Contains(t, expected, actualTool, "Registered tool %s should be in expected list", actualTool)
	}

	// If this fails, update expectedRegisteredTools variable above
	assert.Equal(t, expected, actualTools, "Expected and actual tool lists should match exactly")
}

// TestMCPServerCommunication verifies all tools can be reached through MCP server
func TestMCPServerCommunication(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	// Get expected tools from filesystem
	expectedTools := getExpectedToolsFromFilesystem(t)

	// Simple smoke test: verify all registered tools can be called through MCP server
	// We're not testing the tool logic (that's covered by unit tests)
	// We're only testing MCP server routing and tool registration

	for _, toolName := range expectedTools {
		t.Run(toolName, func(t *testing.T) {
			// Verify tool is registered
			tool := mcputil.GetRegisteredTool(toolName)
			require.NotNil(t, tool, "Tool %s should be registered", toolName)

			// Use the improved HasRequiredParams() method that checks both individual
			// required properties and complex requirements (like RequiresOneOf)
			if tool.HasRequiredParams() {
				// For tools that require parameters, expect controlled validation errors
				err := env.CallToolExpectError(t, toolName, map[string]any{})
				assert.Error(t, err, "Tool %s should return validation error (proving MCP server routing works)", toolName)
				// Validation errors prove the tool was reached through MCP server
				assert.NotEmpty(t, err.Error(), "Tool %s should return non-empty validation error", toolName)
			} else {
				// For tools that work with empty params, expect success
				result := env.CallTool(t, toolName, map[string]any{})
				assert.NotNil(t, result, "Tool %s should return result through MCP server", toolName)
			}
		})
	}
}

// TestSessionManagement tests basic session workflow
func TestSessionManagement(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("SessionTokenGeneration", func(t *testing.T) {
		token := env.GetSessionToken()
		assert.NotEmpty(t, token, "Session token should be generated")
		assert.Greater(t, len(token), 10, "Session token should be substantial length")
	})
}

// TestMCPServerInfrastructure tests the test infrastructure
func TestMCPServerInfrastructure(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("EnvironmentSetup", func(t *testing.T) {
		assert.NotNil(t, env.server, "MCP server should be created")
		assert.NotEmpty(t, env.GetTestDir(), "Test directory should be available")
		assert.DirExists(t, env.GetTestDir(), "Test directory should exist")
	})
}
