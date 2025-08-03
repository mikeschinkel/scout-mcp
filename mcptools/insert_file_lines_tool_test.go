package mcptools_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const InsertFileLinesDirPrefix = "insert-file-lines-tool-test"

// Insert file lines tool result type
type InsertFileLinesResult struct {
	Success    bool   `json:"success"`
	FilePath   string `json:"file_path"`
	LineNumber int    `json:"line_number"`
	Position   string `json:"position"`
	Message    string `json:"message"`
}

type insertFileLinesResultOpts struct {
	ExpectError          bool
	ExpectedErrorMsg     string
	ExpectedFilePath     string
	ExpectedLineNumber   int
	ExpectedPosition     string
	ExpectedContent      string
	ShouldUpdateFile     bool
	ShouldContainText    string
	ShouldNotContainText string
}

func requireInsertFileLinesResult(t *testing.T, result *InsertFileLinesResult, err error, opts insertFileLinesResultOpts) {
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

	if opts.ExpectedLineNumber > 0 {
		assert.Equal(t, opts.ExpectedLineNumber, result.LineNumber, "Line number should match expected")
	}

	if opts.ExpectedPosition != "" {
		assert.Equal(t, opts.ExpectedPosition, result.Position, "Position should match expected")
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

func TestInsertFileLinesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("insert_file_lines")
	require.NotNil(t, tool, "insert_file_lines tool should be registered")

	t.Run("InsertAfterLine_ShouldAddContentAfterSpecifiedLine", func(t *testing.T) {
		tf := testutil.NewTestFixture(InsertFileLinesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("insert-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("insert_test.txt", testutil.FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"filepath":      testFile.Filepath,
			"position":      "after",
			"line_number":   "1",
			"new_content":   "Inserted after line 1",
		})

		result, err := mcputil.GetToolResult[InsertFileLinesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error inserting after line")

		requireInsertFileLinesResult(t, result, err, insertFileLinesResultOpts{
			ExpectedFilePath:   testFile.Filepath,
			ExpectedLineNumber: 1,
			ExpectedPosition:   "after",
			ShouldUpdateFile:   true,
			ExpectedContent:    "Line 1\nInserted after line 1\nLine 2\nLine 3\n",
		})
	})

	t.Run("InsertBeforeLine_ShouldAddContentBeforeSpecifiedLine", func(t *testing.T) {
		tf := testutil.NewTestFixture(InsertFileLinesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("insert-before-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("insert_before_test.txt", testutil.FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"filepath":      testFile.Filepath,
			"position":      "before",
			"line_number":   "2",
			"new_content":   "Inserted before line 2",
		})

		result, err := mcputil.GetToolResult[InsertFileLinesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error inserting before line")

		requireInsertFileLinesResult(t, result, err, insertFileLinesResultOpts{
			ExpectedFilePath:   testFile.Filepath,
			ExpectedLineNumber: 2,
			ExpectedPosition:   "before",
			ShouldUpdateFile:   true,
			ExpectedContent:    "Line 1\nInserted before line 2\nLine 2\nLine 3\n",
		})
	})
}
