package mcptools_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const InsertAtPatternDirPrefix = "insert-at-pattern-tool-test"

// Insert at pattern tool result type
type InsertAtPatternResult struct {
	Success    bool   `json:"success"`
	FilePath   string `json:"file_path"`
	Pattern    string `json:"pattern"`
	Position   string `json:"position"`
	Insertions int    `json:"insertions"`
	Message    string `json:"message"`
}

type insertAtPatternResultOpts struct {
	ExpectError          bool
	ExpectedErrorMsg     string
	ExpectedFilePath     string
	ExpectedPattern      string
	ExpectedPosition     string
	ExpectedInsertions   int
	ExpectedContent      string
	ShouldUpdateFile     bool
	ShouldContainText    string
	ShouldNotContainText string
}

func requireInsertAtPatternResult(t *testing.T, result *InsertAtPatternResult, err error, opts insertAtPatternResultOpts) {
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

	if opts.ExpectedPattern != "" {
		assert.Equal(t, opts.ExpectedPattern, result.Pattern, "Pattern should match expected")
	}

	if opts.ExpectedPosition != "" {
		assert.Equal(t, opts.ExpectedPosition, result.Position, "Position should match expected")
	}

	if opts.ExpectedInsertions > 0 {
		assert.Equal(t, opts.ExpectedInsertions, result.Insertions, "Insertions should match expected")
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

func TestInsertAtPatternTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("insert_at_pattern")
	require.NotNil(t, tool, "insert_at_pattern tool should be registered")

	t.Run("InsertAfterPattern_ShouldAddContentAfterMatchingPattern", func(t *testing.T) {
		tf := testutil.NewTestFixture(InsertAtPatternDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("pattern-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("pattern_test.go", testutil.FileFixtureArgs{
			Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"after_pattern": "func main() {",
			"new_content":   "\n\t// Added comment",
		})

		result, err := getToolResult[InsertAtPatternResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error inserting after pattern",
		)

		requireInsertAtPatternResult(t, result, err, insertAtPatternResultOpts{
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
		tf := testutil.NewTestFixture(InsertAtPatternDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("pattern-before-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("pattern_before_test.go", testutil.FileFixtureArgs{
			Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":  testToken,
			"path":           testFile.Filepath,
			"before_pattern": "func main()",
			"new_content":    "// Main function\n",
		})

		result, err := getToolResult[InsertAtPatternResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error inserting before pattern",
		)

		requireInsertAtPatternResult(t, result, err, insertAtPatternResultOpts{
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
		tf := testutil.NewTestFixture(InsertAtPatternDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("regex-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("regex_pattern_test.go", testutil.FileFixtureArgs{
			Content:     "package main\n\nfunc test() {\n\treturn\n}\n\nfunc another() {\n\treturn\n}\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":  testToken,
			"path":           testFile.Filepath,
			"before_pattern": "func \\w+\\(\\)",
			"new_content":    "// Function comment\n",
			"regex":          true,
		})

		result, err := getToolResult[InsertAtPatternResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with regex pattern",
		)

		requireInsertAtPatternResult(t, result, err, insertAtPatternResultOpts{
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
