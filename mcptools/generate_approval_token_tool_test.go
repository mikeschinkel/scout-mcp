package mcptools_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const GenerateApprovalTokenDirPrefix = "generate-approval-token-tool-test"

// Generate approval token tool result type
type GenerateApprovalTokenResult struct {
	Success   bool   `json:"success"`
	Token     string `json:"token"`
	ExpiresIn string `json:"expires_in"`
	Message   string `json:"message"`
}

type generateApprovalTokenResultOpts struct {
	ExpectError      bool
	ExpectedErrorMsg string
}

func requireGenerateApprovalTokenResult(t *testing.T, result *GenerateApprovalTokenResult, err error, opts generateApprovalTokenResultOpts) {
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

	assert.True(t, result.Success, "Should be successful")
	assert.NotEmpty(t, result.Token, "Token should not be empty")
	assert.Contains(t, result.Token, "mock-approval-token", "Should contain mock token")
	assert.Equal(t, "1 hour", result.ExpiresIn, "Should expire in 1 hour")
	assert.Equal(t, "Approval token generated successfully", result.Message, "Should have success message")
}

func TestGenerateApprovalTokenTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("generate_approval_token")
	require.NotNil(t, tool, "generate_approval_token tool should be registered")

	t.Run("GenerateTokenForFileOperations", func(t *testing.T) {
		tf := testutil.NewTestFixture(GenerateApprovalTokenDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"file_actions": []any{
				mcptools.FileAction{Action: "create", Path: "/test/file1.txt", Purpose: "test creation"},
				mcptools.FileAction{Action: "update", Path: "/test/file2.txt", Purpose: "test update"},
			},
			"operations": []any{"create_file", "update_file"},
		})

		result, err := getToolResult[GenerateApprovalTokenResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error generating approval token",
		)

		requireGenerateApprovalTokenResult(t, result, err, generateApprovalTokenResultOpts{})
	})

	t.Run("GenerateTokenForDeleteOperations", func(t *testing.T) {
		tf := testutil.NewTestFixture(GenerateApprovalTokenDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"file_actions": []any{
				mcptools.FileAction{Action: "delete", Path: "/test/file.txt", Purpose: "test deletion"},
			},
			"operations": []any{"delete_files"},
		})

		result, err := getToolResult[GenerateApprovalTokenResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error generating delete approval token",
		)

		requireGenerateApprovalTokenResult(t, result, err, generateApprovalTokenResultOpts{})
	})

	t.Run("MissingRequiredParameters", func(t *testing.T) {
		tf := testutil.NewTestFixture(GenerateApprovalTokenDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		// Note: The tool treats missing file_actions and operations as empty arrays,
		// which is valid behavior - it generates a token for empty action lists
		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			// Missing file_actions and operations - treated as empty arrays
		})

		result, err := getToolResult[GenerateApprovalTokenResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Tool accepts empty file_actions and operations",
		)

		requireGenerateApprovalTokenResult(t, result, err, generateApprovalTokenResultOpts{})
	})
}
