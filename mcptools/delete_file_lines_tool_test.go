package mcptools_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const DeleteFileLinesDirPrefix = "delete-file-lines-tool-test"

// Delete file lines tool result type
type DeleteFileLinesResult struct {
	Success      bool   `json:"success"`
	FilePath     string `json:"file_path"`
	StartLine    int    `json:"start_line"`
	EndLine      int    `json:"end_line"`
	LinesDeleted int    `json:"lines_deleted"`
	Message      string `json:"message"`
}

type deleteFileLinesResultOpts struct {
	ExpectError          bool
	ExpectedErrorMsg     string
	ExpectedFilePath     string
	ExpectedStartLine    int
	ExpectedEndLine      int
	ExpectedLinesDeleted int
	ExpectedContent      string
	ShouldUpdateFile     bool
	ShouldContainText    string
	ShouldNotContainText string
}

func requireDeleteFileLinesResult(t *testing.T, result *DeleteFileLinesResult, err error, opts deleteFileLinesResultOpts) {
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

	if opts.ExpectedStartLine > 0 {
		assert.Equal(t, opts.ExpectedStartLine, result.StartLine, "Start line should match expected")
	}

	if opts.ExpectedEndLine > 0 {
		assert.Equal(t, opts.ExpectedEndLine, result.EndLine, "End line should match expected")
	}

	if opts.ExpectedLinesDeleted > 0 {
		assert.Equal(t, opts.ExpectedLinesDeleted, result.LinesDeleted, "Lines deleted should match expected")
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

		if opts.ShouldContainText != "" {
			content, readErr := os.ReadFile(opts.ExpectedFilePath)
			require.NoError(t, readErr, "Should be able to read updated file")
			assert.Contains(t, string(content), opts.ShouldContainText, "File should contain expected text")
		}

		if opts.ShouldNotContainText != "" {
			content, readErr := os.ReadFile(opts.ExpectedFilePath)
			require.NoError(t, readErr, "Should be able to read updated file")
			assert.NotContains(t, string(content), opts.ShouldNotContainText, "File should not contain specified text")
		}
	}
}

func TestDeleteFileLinesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("delete_file_lines")
	require.NotNil(t, tool, "delete_file_lines tool should be registered")

	t.Run("DeleteSingleLine_ShouldRemoveSpecifiedLine", func(t *testing.T) {
		tf := testutil.NewTestFixture(DeleteFileLinesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("delete-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("delete_test.txt", testutil.FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"filepath":      testFile.Filepath,
			"start_line":    "3",
			"end_line":      "3",
		})

		result, err := mcputil.GetToolResult[DeleteFileLinesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error deleting single line")

		requireDeleteFileLinesResult(t, result, err, deleteFileLinesResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedStartLine: 3,
			ExpectedEndLine:   3,
			ShouldUpdateFile:  true,
			ExpectedContent:   "Line 1\nLine 2\nLine 4\nLine 5\n",
		})
	})

	t.Run("DeleteMultipleLines_ShouldRemoveSpecifiedLineRange", func(t *testing.T) {
		tf := testutil.NewTestFixture(DeleteFileLinesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("delete-multi-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("delete_multi_test.txt", testutil.FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"filepath":      testFile.Filepath,
			"start_line":    "2",
			"end_line":      "4",
		})

		result, err := mcputil.GetToolResult[DeleteFileLinesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error deleting multiple lines")

		requireDeleteFileLinesResult(t, result, err, deleteFileLinesResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedStartLine: 2,
			ExpectedEndLine:   4,
			ShouldUpdateFile:  true,
			ExpectedContent:   "Line 1\nLine 5\n",
		})
	})
}
