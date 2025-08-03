package mcptools_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const UpdateFileLinesDirPrefix = "update-file-lines-tool-test"

// Update file lines tool result type
type UpdateFileLinesResult struct {
	Success      bool   `json:"success"`
	FilePath     string `json:"file_path"`
	StartLine    int    `json:"start_line"`
	EndLine      int    `json:"end_line"`
	LinesUpdated int    `json:"lines_updated"`
	Message      string `json:"message"`
}

type updateFileLinesResultOpts struct {
	ExpectError          bool
	ExpectedErrorMsg     string
	ExpectedFilePath     string
	ExpectedStartLine    int
	ExpectedEndLine      int
	ExpectedLinesUpdated int
	ExpectedContent      string
	ShouldUpdateFile     bool
	ShouldContainText    string
	ShouldNotContainText string
}

func requireUpdateFileLinesResult(t *testing.T, result *UpdateFileLinesResult, err error, opts updateFileLinesResultOpts) {
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

	if opts.ExpectedLinesUpdated > 0 {
		assert.Equal(t, opts.ExpectedLinesUpdated, result.LinesUpdated, "Lines updated should match expected")
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

func TestUpdateFileLinesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("update_file_lines")
	require.NotNil(t, tool, "update_file_lines tool should be registered")

	t.Run("UpdateSingleLine_ShouldReplaceLineWithNewContent", func(t *testing.T) {
		tf := testutil.NewTestFixture(UpdateFileLinesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("update-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("update_lines_test.txt", testutil.FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"filepath":      testFile.Filepath,
			"start_line":    "2",
			"end_line":      "2",
			"new_content":   "Updated Line 2",
		})

		result, err := mcputil.GetToolResult[UpdateFileLinesResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error updating single line")

		requireUpdateFileLinesResult(t, result, err, updateFileLinesResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedStartLine: 2,
			ExpectedEndLine:   2,
			ShouldUpdateFile:  true,
			ExpectedContent:   "Line 1\nUpdated Line 2\nLine 3\nLine 4\nLine 5\n",
		})
	})

	t.Run("UpdateMultipleLines_ShouldReplaceLineRangeWithNewContent", func(t *testing.T) {
		tf := testutil.NewTestFixture(UpdateFileLinesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("update-multi-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("update_multi_lines_test.txt", testutil.FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"filepath":      testFile.Filepath,
			"start_line":    "2",
			"end_line":      "4",
			"new_content":   "Updated Lines 2-4",
		})

		result, err := mcputil.GetToolResult[UpdateFileLinesResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error updating multiple lines")

		requireUpdateFileLinesResult(t, result, err, updateFileLinesResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedStartLine: 2,
			ExpectedEndLine:   4,
			ShouldUpdateFile:  true,
			ExpectedContent:   "Line 1\nUpdated Lines 2-4\nLine 5\n",
		})
	})

	t.Run("UpdateInvalidLineRange_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := testutil.NewTestFixture(UpdateFileLinesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("error-test-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		// Create a valid file for testing line validation
		testFile := pf.AddFileFixture("test.txt", testutil.FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"filepath":      testFile.Filepath,
			"start_line":    0, // Invalid line number - should be >= 1
			"end_line":      2,
			"new_content":   "Updated content",
		})

		result, err := mcputil.GetToolResult[UpdateFileLinesResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should handle invalid line range")

		requireUpdateFileLinesResult(t, result, err, updateFileLinesResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "start_line must be >= 1",
		})
	})
}
