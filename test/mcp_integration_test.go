package test

import (
	"sort"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedRegisteredTools is the list of tools we expect to be registered
// This must be kept in sync with actual tool registrations
var expectedRegisteredTools = []string{
	"start_session", "get_config", "help", "detect_current_project",
	"read_files", "search_files", "analyze_files", "validate_files",
	"create_file", "update_file", "delete_files",
	"update_file_lines", "delete_file_lines", "insert_file_lines",
	"insert_at_pattern", "replace_pattern",
	"find_file_part", "replace_file_part",
	"request_approval",
}

// TestToolRegistrationCompleteness verifies our expected tool list matches actual registrations
func TestToolRegistrationCompleteness(t *testing.T) {
	// Get actual registered tool names
	actualTools := mcputil.GetRegisteredToolNames()
	sort.Strings(actualTools)

	// Sort expected tools for comparison
	expected := make([]string, len(expectedRegisteredTools))
	copy(expected, expectedRegisteredTools)
	sort.Strings(expected)

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

	// Simple smoke test: verify all registered tools can be called through MCP server
	// We're not testing the tool logic (that's covered by unit tests)
	// We're only testing MCP server routing and tool registration

	for _, toolName := range expectedRegisteredTools {
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
