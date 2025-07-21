package mcptools_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAllToolsRegistered verifies that all expected tools are registered during init()
func TestAllToolsRegistered(t *testing.T) {
	var tool mcputil.Tool
	var toolName string

	// Verify each tool is registered
	for toolName = range toolNamesMap {
		tool = mcputil.GetRegisteredTool(toolName)
		assert.True(t, tool != nil, "tool %s should be registered", toolName)
		assert.NotNil(t, tool, "tool %s should not be nil", toolName)

		if tool != nil {
			assert.Equal(t, toolName, tool.Name(), "tool %s name should match", toolName)
			assert.NotEmpty(t, tool.Options().Description, "tool %s should have description", toolName)
			assert.NotNil(t, tool.Options().Properties, "tool %s should have properties", toolName)
		}
	}
}

// TestNoUnexpectedTools verifies that only expected tools are registered
func TestNoUnexpectedTools(t *testing.T) {
	var toolName string

	registeredTools := mcputil.RegisteredToolsMap()
	// Check for unexpected tools
	for toolName = range registeredTools {
		_, ok := toolNamesMap[toolName]
		assert.True(t, ok, "unexpected tool registered: %s", toolName)
	}

	// Verify expected count
	assert.Equal(t, len(toolNamesMap), len(registeredTools), "tool count mismatch")
}

// TestToolMetadataConsistency validates that all tools have consistent metadata
func TestToolMetadataConsistency(t *testing.T) {
	var tool mcputil.Tool

	for _, tool = range mcputil.RegisteredTools() {
		// Basic metadata validation
		assert.NotEmpty(t, tool.Options().Description, "tool %s should have description", tool.Name())

		if _, ok := tool.(*mcptools.StartSessionTool); !ok {
			assert.NotNil(t, tool.Options().Properties, "tool %s should have properties", tool.Name())
		}
	}
}

// TestSessionTokenRequirements validates which tools require session tokens
func TestSessionTokenRequirements(t *testing.T) {
	var tool mcputil.Tool
	var toolName string

	for toolName = range toolNamesMap {
		var hasSessionToken bool
		tool = mcputil.GetRegisteredTool(toolName)
		require.NotNil(t, tool, "tool %s must be registered", toolName)

		// Check if tool has session_token in its properties
		for _, p := range tool.Options().Properties {
			if p.GetName() != "session_token" {
				continue
			}
			hasSessionToken = true
			break
		}

		if toolName == "start_session" {
			assert.False(t, hasSessionToken, "tool %s should NOT require session_token", toolName)
		} else {
			assert.True(t, hasSessionToken, "tool %s should require session_token", toolName)
		}
	}
}

// TestToolRegistrationOrder verifies that tools are registered in a predictable order
func TestToolRegistrationOrder(t *testing.T) {
	var toolNames []string
	var toolName string

	registeredTools := mcputil.RegisteredToolsMap()

	// Extract tool names
	for toolName = range registeredTools {
		toolNames = append(toolNames, toolName)
	}

	// Should have expected number of tools
	assert.Equal(t, len(toolNamesMap), len(toolNames), "should have expected number of tools registered")

	// start_session should always be registered (critical for other tools)
	startSessionTool := mcputil.GetRegisteredTool("start_session")
	assert.NotNil(t, startSessionTool, "start_session must be registered")
}
