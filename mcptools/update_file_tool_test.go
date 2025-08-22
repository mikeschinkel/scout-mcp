package mcptools_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const UpdateFileDirPrefix = "update-file-tool-test"

// Update file tool result type
type UpdateFileResult struct {
	Success  bool   `json:"success"`
	FilePath string `json:"file_path"`
	OldSize  int64  `json:"old_size"`
	NewSize  int64  `json:"new_size"`
	Message  string `json:"message"`
}

type updateFileResultOpts struct {
	ExpectError      bool
	ExpectedErrorMsg string
	ExpectedFilePath string
	ShouldUpdateFile bool
	ExpectedContent  string
}

func requireUpdateFileResult(t *testing.T, result *UpdateFileResult, err error, opts updateFileResultOpts) {
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

	assert.True(t, result.Success, "Operation should be successful")

	if opts.ExpectedFilePath != "" {
		assert.Equal(t, opts.ExpectedFilePath, result.FilePath, "File path should match expected")
	}

	// Check file system side effects
	if opts.ShouldUpdateFile && opts.ExpectedFilePath != "" {
		_, err := os.Stat(opts.ExpectedFilePath)
		assert.NoError(t, err, "File should exist on disk: %s", opts.ExpectedFilePath)

		if opts.ExpectedContent != "" {
			content, readErr := os.ReadFile(opts.ExpectedFilePath)
			require.NoError(t, readErr, "Should be able to read updated file")
			assert.Equal(t, opts.ExpectedContent, string(content), "File content should match expected")
		}
	}
}

func TestUpdateFileTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("update_file")
	require.NotNil(t, tool, "update_file tool should be registered")

	t.Run("UpdateExistingFile_ShouldReplaceFileContent", func(t *testing.T) {
		tf := fsfix.NewRootFixture(UpdateFileDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("update-project", nil)
		testFile := pf.AddFileFixture("update_me.txt", &fsfix.FileFixtureArgs{
			Content: "Original content",
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"filepath":      testFile.Filepath,
			"new_content":   "Updated content",
		})

		result, err := mcputil.GetToolResult[UpdateFileResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error updating file")

		requireUpdateFileResult(t, result, err, updateFileResultOpts{
			ShouldUpdateFile: true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedContent:  "Updated content",
		})
	})

	t.Run("UpdateNonexistentFile_ShouldReturnError", func(t *testing.T) {
		tf := fsfix.NewRootFixture(UpdateFileDirPrefix)
		defer tf.Cleanup()
		// Add a missing file for error testing
		nonexistentFile := tf.AddFileFixture("does-not-exist.txt", &fsfix.FileFixtureArgs{
			Missing: true,
		})
		tf.Setup(t)

		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"filepath":      nonexistentFile.Filepath,
			"new_content":   "This should fail",
		})

		result, err := mcputil.GetToolResult[UpdateFileResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should handle nonexistent file")

		requireUpdateFileResult(t, result, err, updateFileResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "file does not exist",
		})
	})
}
