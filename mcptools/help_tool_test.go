package mcptools_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const HelpDirPrefix = "scout-help-tool-test"

// Tool help tool result type
type HelpResult struct {
	Tool            string                `json:"tool"`
	Content         string                `json:"content"`
	Type            string                `json:"type"`
	PayloadTypeName string                `json:"payload_type_name"`
	Payload         *mcptools.HelpPayload `json:"payload"`
}

func TestScoutHelpTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("help")
	require.NotNil(t, tool, "help tool should be registered")

	t.Run("GetFullDocumentationWithScoutContent", func(t *testing.T) {
		tf := testutil.NewTestFixture(HelpDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": mcputil.TestToken,
		})

		result, err := mcputil.GetToolResult[HelpResult](
			mcputil.CallResult(mcputil.CallTool(tool, req)),
			"Should not error getting full documentation with Scout content",
		)

		require.NoError(t, err, "Should not have error")
		require.NotNil(t, result, "Result should not be nil")
		assert.NotEmpty(t, result.Content, "Help content should not be empty")
		assert.Equal(t, "full_documentation", result.Type, "Type should be full_documentation")

		// Scout-specific assertions
		assert.NotNil(t, result.Payload, "Should have Scout-specific payload")
		assert.NotEmpty(t, result.Payload.ServerSpecificHelp, "Should have Scout-specific help content")
		assert.Contains(t, result.Payload.ServerSpecificHelp, "Scout MCP", "Scout help should mention Scout MCP")
	})

	t.Run("GetSpecificHelpForScoutTool", func(t *testing.T) {
		tf := testutil.NewTestFixture(HelpDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": mcputil.TestToken,
			"tool":          "read_files",
		})

		result, err := mcputil.GetToolResult[HelpResult](
			mcputil.CallResult(mcputil.CallTool(tool, req)),
			"Should not error getting specific Scout tool help",
		)

		require.NoError(t, err, "Should not have error")
		require.NotNil(t, result, "Result should not be nil")
		assert.NotEmpty(t, result.Content, "Help content should not be empty")
		assert.Contains(t, result.Content, "read_files", "Help should contain read_files tool info")
		assert.Equal(t, "read_files", result.Tool, "Tool should match requested")
		assert.Equal(t, "tool_specific", result.Type, "Type should be tool_specific")

		// Scout-specific payload should still be present
		assert.NotNil(t, result.Payload, "Should have Scout-specific payload")
	})
}
