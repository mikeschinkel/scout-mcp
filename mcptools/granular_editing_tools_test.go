package mcptools_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const GranularEditingDirPrefix = "granular-editing-test"

type granularEditingResultOpts struct {
	ExpectError          bool
	ExpectedErrorMsg     string
	ExpectedFilePath     string
	ExpectedStartLine    int
	ExpectedEndLine      int
	ExpectedLineNumber   int
	ExpectedPosition     string
	ExpectedPattern      string
	ExpectedReplacement  string
	ExpectedReplacements int
	ExpectedContent      string
	ShouldUpdateFile     bool
	ShouldContainText    string
	ShouldNotContainText string
}

func requireGranularEditingResult(t *testing.T, result *map[string]any, err error, opts granularEditingResultOpts) {
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

	// Check basic result structure
	if success, hasSuccess := (*result)["success"].(bool); hasSuccess {
		assert.True(t, success, "Operation should be successful")
	}

	if opts.ExpectedFilePath != "" {
		if filePath, hasFilePath := (*result)["file_path"].(string); hasFilePath {
			assert.Equal(t, opts.ExpectedFilePath, filePath, "File path should match expected")
		}
	}

	// Check line-specific fields
	if opts.ExpectedStartLine > 0 {
		if startLine, hasStartLine := (*result)["start_line"].(float64); hasStartLine {
			assert.Equal(t, float64(opts.ExpectedStartLine), startLine, "Start line should match expected")
		}
	}

	if opts.ExpectedEndLine > 0 {
		if endLine, hasEndLine := (*result)["end_line"].(float64); hasEndLine {
			assert.Equal(t, float64(opts.ExpectedEndLine), endLine, "End line should match expected")
		}
	}

	if opts.ExpectedLineNumber > 0 {
		if lineNumber, hasLineNumber := (*result)["line_number"].(float64); hasLineNumber {
			assert.Equal(t, float64(opts.ExpectedLineNumber), lineNumber, "Line number should match expected")
		}
	}

	if opts.ExpectedPosition != "" {
		if position, hasPosition := (*result)["position"].(string); hasPosition {
			assert.Equal(t, opts.ExpectedPosition, position, "Position should match expected")
		}
	}

	// Check pattern-specific fields
	if opts.ExpectedPattern != "" {
		if pattern, hasPattern := (*result)["pattern"].(string); hasPattern {
			assert.Equal(t, opts.ExpectedPattern, pattern, "Pattern should match expected")
		}
	}

	if opts.ExpectedReplacement != "" {
		if replacement, hasReplacement := (*result)["replacement"].(string); hasReplacement {
			assert.Equal(t, opts.ExpectedReplacement, replacement, "Replacement should match expected")
		}
	}

	if opts.ExpectedReplacements > 0 {
		if replacementCount, hasReplacementCount := (*result)["replacement_count"].(float64); hasReplacementCount {
			assert.Equal(t, float64(opts.ExpectedReplacements), replacementCount, "Replacement count should match expected")
		}
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
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("update-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("update_lines_test.txt", FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      testFile.Filepath,
			"start_line":    "2",
			"end_line":      "2",
			"new_content":   "Updated Line 2",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error updating single line",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedStartLine: 2,
			ExpectedEndLine:   2,
			ShouldUpdateFile:  true,
			ExpectedContent:   "Line 1\nUpdated Line 2\nLine 3\nLine 4\nLine 5\n",
		})
	})

	t.Run("UpdateMultipleLines_ShouldReplaceLineRangeWithNewContent", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("update-multi-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("update_multi_lines_test.txt", FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      testFile.Filepath,
			"start_line":    "2",
			"end_line":      "4",
			"new_content":   "Updated Lines 2-4",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error updating multiple lines",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedStartLine: 2,
			ExpectedEndLine:   4,
			ShouldUpdateFile:  true,
			ExpectedContent:   "Line 1\nUpdated Lines 2-4\nLine 5\n",
		})
	})

	t.Run("UpdateInvalidLineRange_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("error-test-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		// Create a valid file for testing line validation
		testFile := pf.AddFileFixture("test.txt", FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      testFile.Filepath,
			"start_line":    0, // Invalid line number - should be >= 1
			"end_line":      2,
			"new_content":   "Updated content",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle invalid line range",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "start_line must be >= 1",
		})
	})
}

func TestInsertFileLinesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("insert_file_lines")
	require.NotNil(t, tool, "insert_file_lines tool should be registered")

	t.Run("InsertAfterLine_ShouldAddContentAfterSpecifiedLine", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("insert-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("insert_test.txt", FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      testFile.Filepath,
			"position":      "after",
			"line_number":   "1",
			"new_content":   "Inserted after line 1",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error inserting after line",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:   testFile.Filepath,
			ExpectedLineNumber: 1,
			ExpectedPosition:   "after",
			ShouldUpdateFile:   true,
			ExpectedContent:    "Line 1\nInserted after line 1\nLine 2\nLine 3\n",
		})
	})

	t.Run("InsertBeforeLine_ShouldAddContentBeforeSpecifiedLine", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("insert-before-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("insert_before_test.txt", FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      testFile.Filepath,
			"position":      "before",
			"line_number":   "2",
			"new_content":   "Inserted before line 2",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error inserting before line",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:   testFile.Filepath,
			ExpectedLineNumber: 2,
			ExpectedPosition:   "before",
			ShouldUpdateFile:   true,
			ExpectedContent:    "Line 1\nInserted before line 2\nLine 2\nLine 3\n",
		})
	})
}

func TestDeleteFileLinesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("delete_file_lines")
	require.NotNil(t, tool, "delete_file_lines tool should be registered")

	t.Run("DeleteSingleLine_ShouldRemoveSpecifiedLine", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("delete-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("delete_test.txt", FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      testFile.Filepath,
			"start_line":    "3",
			"end_line":      "3",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error deleting single line",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedStartLine: 3,
			ExpectedEndLine:   3,
			ShouldUpdateFile:  true,
			ExpectedContent:   "Line 1\nLine 2\nLine 4\nLine 5\n",
		})
	})

	t.Run("DeleteMultipleLines_ShouldRemoveSpecifiedLineRange", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("delete-multi-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("delete_multi_test.txt", FileFixtureArgs{
			Content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      testFile.Filepath,
			"start_line":    "2",
			"end_line":      "4",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error deleting multiple lines",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedStartLine: 2,
			ExpectedEndLine:   4,
			ShouldUpdateFile:  true,
			ExpectedContent:   "Line 1\nLine 5\n",
		})
	})
}

func TestInsertAtPatternTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("insert_at_pattern")
	require.NotNil(t, tool, "insert_at_pattern tool should be registered")

	t.Run("InsertAfterPattern_ShouldAddContentAfterMatchingPattern", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("pattern-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("pattern_test.go", FileFixtureArgs{
			Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"after_pattern": "func main() {",
			"new_content":   "\n\t// Added comment",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error inserting after pattern",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedPattern:   "func main() {",
			ShouldUpdateFile:  true,
			ShouldContainText: "\t// Added comment",
		})

		// Verify the function is still there
		content, readErr := os.ReadFile(testFile.Filepath)
		require.NoError(t, readErr, "Should be able to read updated file")
		assert.Contains(t, string(content), "func main() {", "Function should still be there")
	})

	t.Run("InsertBeforePattern_ShouldAddContentBeforeMatchingPattern", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("pattern-before-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("pattern_before_test.go", FileFixtureArgs{
			Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":  tf.token,
			"path":           testFile.Filepath,
			"before_pattern": "func main()",
			"new_content":    "// Main function\n",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error inserting before pattern",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedPattern:   "func main()",
			ShouldUpdateFile:  true,
			ShouldContainText: "// Main function",
		})

		// Verify the function is still there
		content, readErr := os.ReadFile(testFile.Filepath)
		require.NoError(t, readErr, "Should be able to read updated file")
		assert.Contains(t, string(content), "func main()", "Function should still be there")
	})

	t.Run("RegexPattern_ShouldMatchUsingRegularExpression", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("regex-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("regex_pattern_test.go", FileFixtureArgs{
			Content:     "package main\n\nfunc test() {\n\treturn\n}\n\nfunc another() {\n\treturn\n}\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":  tf.token,
			"path":           testFile.Filepath,
			"before_pattern": "func \\w+\\(\\)",
			"new_content":    "// Function comment\n",
			"regex":          true,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with regex pattern",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:  testFile.Filepath,
			ExpectedPattern:   "func \\w+\\(\\)",
			ShouldUpdateFile:  true,
			ShouldContainText: "// Function comment",
		})

		// Verify the first function is still there
		content, readErr := os.ReadFile(testFile.Filepath)
		require.NoError(t, readErr, "Should be able to read updated file")
		assert.Contains(t, string(content), "func test()", "First function should still be there")
	})
}

func TestReplacePatternTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("replace_pattern")
	require.NotNil(t, tool, "replace_pattern tool should be registered")

	t.Run("SimpleTextReplace_ShouldReplaceAllOccurrencesOfPattern", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_test.txt", FileFixtureArgs{
			Content:     "Hello old world\nThis is old content\nold values everywhere",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":   tf.token,
			"path":            testFile.Filepath,
			"pattern":         "old",
			"replacement":     "new",
			"all_occurrences": true,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error replacing text",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:     testFile.Filepath,
			ExpectedPattern:      "old",
			ExpectedReplacement:  "new",
			ExpectedReplacements: 3,
			ShouldUpdateFile:     true,
			ExpectedContent:      "Hello new world\nThis is new content\nnew values everywhere",
		})
	})

	t.Run("ReplaceFirstOnly_ShouldReplaceOnlyFirstOccurrence", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-first-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_first_test.txt", FileFixtureArgs{
			Content:     "test value\ntest again\ntest final",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":   tf.token,
			"path":            testFile.Filepath,
			"pattern":         "test",
			"replacement":     "demo",
			"all_occurrences": false,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error replacing first occurrence",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:     testFile.Filepath,
			ExpectedPattern:      "test",
			ExpectedReplacement:  "demo",
			ExpectedReplacements: 1,
			ShouldUpdateFile:     true,
			ExpectedContent:      "demo value\ntest again\ntest final",
		})
	})

	t.Run("RegexReplace_ShouldUseRegularExpressionForMatching", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("regex-replace-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("regex_replace_test.txt", FileFixtureArgs{
			Content:     "func functionOne() {\n\treturn\n}\n\nfunc functionTwo(param string) {\n\treturn\n}\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":   tf.token,
			"path":            testFile.Filepath,
			"pattern":         "func (\\w+)\\(",
			"replacement":     "function $1(",
			"regex":           true,
			"all_occurrences": true,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with regex replacement",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectedFilePath:     testFile.Filepath,
			ExpectedPattern:      "func (\\w+)\\(",
			ExpectedReplacement:  "function $1(",
			ExpectedReplacements: 2,
			ShouldUpdateFile:     true,
		})

		// Verify both functions were replaced
		content, readErr := os.ReadFile(testFile.Filepath)
		require.NoError(t, readErr, "Should be able to read updated file")
		assert.Contains(t, string(content), "function functionOne(", "Should replace functionOne")
		assert.Contains(t, string(content), "function functionTwo(", "Should replace functionTwo")
	})

	t.Run("InvalidRegex_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := NewTestFixture(GranularEditingDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("regex-error-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		// Create a valid file for testing regex validation
		testFile := pf.AddFileFixture("test.txt", FileFixtureArgs{
			Content:     "some test content",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"pattern":       "[invalid regex",
			"replacement":   "replacement",
			"regex":         true,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle invalid regex",
		)

		requireGranularEditingResult(t, result, err, granularEditingResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "invalid regex pattern",
		})
	})
}
