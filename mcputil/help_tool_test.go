package mcputil_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const HelpDirPrefix = "tool-help-tool-test"

// Tool help tool result type
type HelpResult struct {
	Tool    string `json:"tool"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type helpToolResultOpts struct {
	ExpectError           bool
	ExpectedErrorMsg      string
	ExpectHelpContent     bool
	ExpectSpecificContent string
	ExpectedTool          string
	ExpectedType          string
}

func requireHelpResult(t *testing.T, result *HelpResult, err error, opts helpToolResultOpts) {
	t.Helper()

	if opts.ExpectError {
		require.Error(t, err, "Should have error")
		if opts.ExpectedErrorMsg != "" {
			assert.Contains(t, err.Error(), opts.ExpectedErrorMsg, "Error should contain expected message")
		}
		return
	}

	require.NoError(t, err, "Should not have error")
	require.NotNil(t, result, "Result should not be nil")
	assert.NotEmpty(t, result.Content, "Help content should not be empty")

	if opts.ExpectedTool != "" {
		assert.Equal(t, opts.ExpectedTool, result.Tool, "Tool should match expected")
	}

	if opts.ExpectedType != "" {
		assert.Equal(t, opts.ExpectedType, result.Type, "Type should match expected")
	}

	if opts.ExpectSpecificContent != "" {
		assert.Contains(t, result.Content, opts.ExpectSpecificContent, "Help should contain expected content")
	}

	if opts.ExpectHelpContent {
		// Basic checks for help content structure
		assert.Contains(t, result.Content, "MCP Tools", "Help should mention MCP Tools")
	}
}

func TestHelpTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("help")
	require.NotNil(t, tool, "help tool should be registered")

	t.Run("GetFullDocumentation", func(t *testing.T) {
		tf := fsfix.NewRootFixture(HelpDirPrefix)
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
			"Should not error getting full documentation",
		)

		requireHelpResult(t, result, err, helpToolResultOpts{
			ExpectHelpContent: true,
			ExpectedType:      "full_documentation",
		})
	})

	t.Run("GetSpecificHelp", func(t *testing.T) {
		tf := fsfix.NewRootFixture(HelpDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": mcputil.TestToken,
			"tool":          "help",
		})

		result, err := mcputil.GetToolResult[HelpResult](
			mcputil.CallResult(mcputil.CallTool(tool, req)),
			"Should not error getting specific tool help",
		)

		requireHelpResult(t, result, err, helpToolResultOpts{
			ExpectSpecificContent: "help",
			ExpectedTool:          "help",
			ExpectedType:          "tool_specific",
		})
	})

	t.Run("GetHelpForNonExistentTool", func(t *testing.T) {
		tf := fsfix.NewRootFixture(HelpDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": mcputil.TestToken,
			"tool":          "nonexistent_tool",
		})

		result, err := mcputil.GetToolResult[HelpResult](
			mcputil.CallResult(mcputil.CallTool(tool, req)),
			"Should not error when tool not found",
		)

		requireHelpResult(t, result, err, helpToolResultOpts{
			ExpectedTool: "nonexistent_tool",
			ExpectedType: "tool_specific",
		})
		// The tool should return helpful message about available tools
	})
}
