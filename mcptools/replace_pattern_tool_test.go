package mcptools_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ReplacePatternDirPrefix = "replace-pattern-tool-test"

// Replace pattern tool result type
type ReplacePatternResult struct {
	Success          bool   `json:"success"`
	FilePath         string `json:"file_path"`
	Pattern          string `json:"pattern"`
	Replacement      string `json:"replacement"`
	ReplacementCount int    `json:"replacement_count"`
	Message          string `json:"message"`
}

type replacePatternResultOpts struct {
	ExpectError              bool
	ExpectedErrorMsg         string
	ExpectedFilePath         string
	ExpectedPattern          string
	ExpectedReplacement      string
	ExpectedReplacementCount int
	ExpectedContent          string
	ShouldUpdateFile         bool
	ShouldContainText        string
	ShouldNotContainText     string
}

func requireReplacePatternResult(t *testing.T, result *ReplacePatternResult, err error, opts replacePatternResultOpts) {
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

	if opts.ExpectedReplacement != "" {
		assert.Equal(t, opts.ExpectedReplacement, result.Replacement, "Replacement should match expected")
	}

	if opts.ExpectedReplacementCount > 0 {
		assert.Equal(t, opts.ExpectedReplacementCount, result.ReplacementCount, "Replacement count should match expected")
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

func TestReplacePatternTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("replace_pattern")
	require.NotNil(t, tool, "replace_pattern tool should be registered")

	t.Run("SimpleTextReplace_ShouldReplaceAllOccurrencesOfPattern", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReplacePatternDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_test.txt", testutil.FileFixtureArgs{
			Content:     "Hello old world\nThis is old content\nold values everywhere",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":   testToken,
			"path":            testFile.Filepath,
			"pattern":         "old",
			"replacement":     "new",
			"all_occurrences": true,
		})

		result, err := mcputil.GetToolResult[ReplacePatternResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error replacing text")

		requireReplacePatternResult(t, result, err, replacePatternResultOpts{
			ExpectedFilePath:         testFile.Filepath,
			ExpectedPattern:          "old",
			ExpectedReplacement:      "new",
			ExpectedReplacementCount: 3,
			ShouldUpdateFile:         true,
			ExpectedContent:          "Hello new world\nThis is new content\nnew values everywhere",
		})
	})

	t.Run("ReplaceFirstOnly_ShouldReplaceOnlyFirstOccurrence", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReplacePatternDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-first-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_first_test.txt", testutil.FileFixtureArgs{
			Content:     "test value\ntest again\ntest final",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":   testToken,
			"path":            testFile.Filepath,
			"pattern":         "test",
			"replacement":     "demo",
			"all_occurrences": false,
		})

		result, err := mcputil.GetToolResult[ReplacePatternResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error replacing first occurrence")

		requireReplacePatternResult(t, result, err, replacePatternResultOpts{
			ExpectedFilePath:         testFile.Filepath,
			ExpectedPattern:          "test",
			ExpectedReplacement:      "demo",
			ExpectedReplacementCount: 1,
			ShouldUpdateFile:         true,
			ExpectedContent:          "demo value\ntest again\ntest final",
		})
	})

	t.Run("RegexReplace_ShouldUseRegularExpressionForMatching", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReplacePatternDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("regex-replace-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("regex_replace_test.txt", testutil.FileFixtureArgs{
			Content:     "func functionOne() {\n\treturn\n}\n\nfunc functionTwo(param string) {\n\treturn\n}\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":   testToken,
			"path":            testFile.Filepath,
			"pattern":         "func (\\w+)\\(",
			"replacement":     "function $1(",
			"regex":           true,
			"all_occurrences": true,
		})

		result, err := mcputil.GetToolResult[ReplacePatternResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error with regex replacement")

		requireReplacePatternResult(t, result, err, replacePatternResultOpts{
			ExpectedFilePath:         testFile.Filepath,
			ExpectedPattern:          "func (\\w+)\\(",
			ExpectedReplacement:      "function $1(",
			ExpectedReplacementCount: 2,
			ShouldUpdateFile:         true,
		})

		// Verify both functions were replaced
		content, readErr := os.ReadFile(testFile.Filepath)
		require.NoError(t, readErr, "Should be able to read updated file")
		assert.Contains(t, string(content), "function functionOne(", "Should replace functionOne")
		assert.Contains(t, string(content), "function functionTwo(", "Should replace functionTwo")
	})

	t.Run("InvalidRegex_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReplacePatternDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("regex-error-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		// Create a valid file for testing regex validation
		testFile := pf.AddFileFixture("test.txt", testutil.FileFixtureArgs{
			Content:     "some test content",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"pattern":       "[invalid regex",
			"replacement":   "replacement",
			"regex":         true,
		})

		result, err := mcputil.GetToolResult[ReplacePatternResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should handle invalid regex")

		requireReplacePatternResult(t, result, err, replacePatternResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "invalid regex pattern",
		})
	})
}
