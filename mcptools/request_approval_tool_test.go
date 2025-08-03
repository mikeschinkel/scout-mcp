package mcptools_test

import (
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const RequestApprovalDirPrefix = "request-approval-tool-test"

// Request approval tool result type
type RequestApprovalResult struct {
	Status         string   `json:"status"`
	Operation      string   `json:"operation"`
	RiskLevel      string   `json:"risk_level"`
	FilesAffected  int      `json:"files_affected"`
	Files          []string `json:"files"`
	ImpactSummary  string   `json:"impact_summary"`
	PreviewContent string   `json:"preview_content"`
	Message        string   `json:"message"`
	Note           string   `json:"note"`
}

type requestApprovalResultOpts struct {
	ExpectError       bool
	ExpectedErrorMsg  string
	ExpectedOperation string
	ExpectedRiskLevel string
	ExpectedFiles     int
}

func requireRequestApprovalResult(t *testing.T, result *RequestApprovalResult, err error, opts requestApprovalResultOpts) {
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

	assert.Equal(t, "approval_requested", result.Status, "Status should be approval_requested")
	assert.Equal(t, "Approval request logged for manual review", result.Message, "Should have correct message")
	assert.NotEmpty(t, result.Note, "Should have note about being a stub implementation")

	if opts.ExpectedOperation != "" {
		assert.Equal(t, opts.ExpectedOperation, result.Operation, "Operation should match expected")
	}

	if opts.ExpectedRiskLevel != "" {
		assert.Equal(t, opts.ExpectedRiskLevel, result.RiskLevel, "Risk level should match expected")
	}

	if opts.ExpectedFiles > 0 {
		assert.Equal(t, opts.ExpectedFiles, result.FilesAffected, "Files affected should match expected")
		assert.Len(t, result.Files, opts.ExpectedFiles, "Files array should have expected length")
	}
}

func TestRequestApprovalTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("request_approval")
	require.NotNil(t, tool, "request_approval tool should be registered")

	t.Run("RequestBasicApproval", func(t *testing.T) {
		tf := testutil.NewTestFixture(RequestApprovalDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token":   testToken,
			"operation":       "create_file",
			"files":           []any{filepath.Join(tf.TempDir(), "new_file.txt")},
			"impact_summary":  "Creating a new configuration file",
			"risk_level":      "low",
			"preview_content": "# Configuration\nport: 8080\nhost: localhost",
		})

		result, err := mcputil.GetToolResult[RequestApprovalResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error requesting approval")

		requireRequestApprovalResult(t, result, err, requestApprovalResultOpts{
			ExpectedOperation: "create_file",
			ExpectedRiskLevel: "low",
			ExpectedFiles:     1,
		})
	})

	t.Run("RequestHighRiskApproval", func(t *testing.T) {
		tf := testutil.NewTestFixture(RequestApprovalDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token":   testToken,
			"operation":       "delete_files",
			"files":           []any{filepath.Join(tf.TempDir(), "important_file.txt")},
			"impact_summary":  "Deleting critical system file",
			"risk_level":      "high",
			"preview_content": "This operation will permanently delete the file",
		})

		result, err := mcputil.GetToolResult[RequestApprovalResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error requesting high-risk approval")

		requireRequestApprovalResult(t, result, err, requestApprovalResultOpts{
			ExpectedOperation: "delete_files",
			ExpectedRiskLevel: "high",
			ExpectedFiles:     1,
		})
	})

	t.Run("InvalidRiskLevel", func(t *testing.T) {
		tf := testutil.NewTestFixture(RequestApprovalDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		// Note: request_approval tool is currently a stub implementation
		// that doesn't validate risk levels yet
		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token":  testToken,
			"operation":      "update_file",
			"files":          []any{filepath.Join(tf.TempDir(), "test.txt")},
			"impact_summary": "Updating file",
			"risk_level":     "invalid_level",
		})

		result, err := mcputil.GetToolResult[RequestApprovalResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Stub implementation doesn't validate risk levels yet")

		requireRequestApprovalResult(t, result, err, requestApprovalResultOpts{
			ExpectedOperation: "update_file",
			ExpectedRiskLevel: "invalid_level",
			ExpectedFiles:     1,
		})
	})
}
