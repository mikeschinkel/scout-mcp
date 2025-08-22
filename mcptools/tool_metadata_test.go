package mcptools_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ToolMetadataDirPrefix = "tool-metadata-test"

// Test tool metadata and registration
func TestToolMetadata(t *testing.T) {
	expectedTools := toolNamesMap

	// Get all registered tools
	registeredTools := mcputil.RegisteredToolsMap()

	t.Run("AllExpectedToolsRegistered", func(t *testing.T) {
		tf := fsfix.NewRootFixture(ToolMetadataDirPrefix)
		defer tf.Cleanup()
		tf.Setup(t)

		for expectedTool := range expectedTools {
			_, exists := registeredTools[expectedTool]
			assert.True(t, exists, "Expected tool %s should be registered", expectedTool)
		}
	})

	t.Run("NoUnexpectedTools", func(t *testing.T) {
		tf := fsfix.NewRootFixture(ToolMetadataDirPrefix)
		defer tf.Cleanup()
		tf.Setup(t)

		for registeredTool := range registeredTools {
			_, expected := expectedTools[registeredTool]
			assert.True(t, expected, "Unexpected tool %s is registered", registeredTool)
		}
	})

	t.Run("ToolCount", func(t *testing.T) {
		tf := fsfix.NewRootFixture(ToolMetadataDirPrefix)
		defer tf.Cleanup()
		tf.Setup(t)

		assert.Equal(t, len(expectedTools), len(registeredTools),
			"Number of registered tools should match expected count")
	})

	t.Run("ToolBasicProperties", func(t *testing.T) {
		tf := fsfix.NewRootFixture(ToolMetadataDirPrefix)
		defer tf.Cleanup()
		tf.Setup(t)

		for toolName := range expectedTools {
			tool := mcputil.GetRegisteredTool(toolName)
			require.NotNil(t, tool, "Tool %s should be accessible", toolName)

			assert.Equal(t, toolName, tool.Name(), "Tool name should match")
			assert.NotEmpty(t, tool.Options().Description, "Tool %s should have description", toolName)

			// All tools except start_session should have session_token property
			if toolName != "start_session" {
				hasSessionToken := false
				for _, prop := range tool.Options().Properties {
					if prop.GetName() == "session_token" {
						hasSessionToken = true
						break
					}
				}
				assert.True(t, hasSessionToken, "Tool %s should have session_token property", toolName)
			}
		}
	})
}
