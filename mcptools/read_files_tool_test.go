package mcptools_test

import (
	"fmt"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ReadFilesDirPrefix = "read-files-tool-test"

// Read files tool result type
type ReadFilesResult struct {
	Files []struct {
		Path    string `json:"path"`
		Name    string `json:"name"`
		Size    int64  `json:"size"`
		Content string `json:"content"`
	} `json:"files"`
	TotalFiles int    `json:"total_files"`
	Summary    string `json:"summary"`
	Errors     []any  `json:"errors,omitempty"`
}

type readFilesResultOpts struct {
	ExpectError        bool
	ExpectedErrorMsg   string
	ExpectPartialError bool
	ExpectFiles        int
	MinFiles           int
	ExpectedContent    string
	ExpectedContents   []string
	ExpectedSize       int64
}

func requireReadFilesResult(t *testing.T, result *ReadFilesResult, err error, opts readFilesResultOpts) {
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

	// Check files count
	if opts.ExpectFiles > 0 {
		assert.Len(t, result.Files, opts.ExpectFiles, "Should have expected number of files")
	}
	if opts.MinFiles > 0 {
		assert.GreaterOrEqual(t, len(result.Files), opts.MinFiles, "Should have at least minimum number of files")
	}

	// Check expected content
	if opts.ExpectedContent != "" {
		found := false
		for _, file := range result.Files {
			if file.Content == opts.ExpectedContent {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find expected content: %s", opts.ExpectedContent)
	}

	// Check multiple expected contents
	if len(opts.ExpectedContents) > 0 {
		foundContents := make(map[string]bool)
		for _, file := range result.Files {
			foundContents[file.Content] = true
		}
		for _, expectedContent := range opts.ExpectedContents {
			assert.True(t, foundContents[expectedContent], "Should find expected content: %s", expectedContent)
		}
	}

	// Check for partial errors
	if opts.ExpectPartialError {
		assert.Greater(t, len(result.Errors), 0, "Should have at least one error")
		if opts.ExpectedErrorMsg != "" {
			found := false
			for _, errorInterface := range result.Errors {
				if errorStr, ok := errorInterface.(string); ok && errorStr != "" {
					if assert.Contains(t, errorStr, opts.ExpectedErrorMsg, "Error should contain expected message") {
						found = true
						break
					}
				}
			}
			assert.True(t, found, "Should find error message containing: %s", opts.ExpectedErrorMsg)
		}
	}
}

func TestReadFilesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("read_files")
	require.NotNil(t, tool, "read_files tool should be registered")

	t.Run("ReadSingleFile_ShouldReturnFileContentAndMetadata", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReadFilesDirPrefix)
		defer tf.Cleanup()

		// Create a test file with known content
		pf := tf.AddProjectFixture("test-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("test.txt", testutil.FileFixtureArgs{
			Content:     "Hello, World!",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"paths":         []any{testFile.Filepath},
		})

		result, err := mcputil.GetToolResult[ReadFilesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error reading single file")

		requireReadFilesResult(t, result, err, readFilesResultOpts{
			ExpectFiles:     1,
			ExpectedContent: "Hello, World!",
		})
	})

	t.Run("ReadMultipleFiles_ShouldReturnAllFileContents", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReadFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("test-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf1 := pf.AddFileFixture("file1.txt", testutil.FileFixtureArgs{
			Content:     "Content 1",
			Permissions: 0644,
		})
		pf2 := pf.AddFileFixture("file2.txt", testutil.FileFixtureArgs{
			Content:     "Content 2",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"paths": []any{
				pf1.Filepath,
				pf2.Filepath,
			},
		})

		result, err := mcputil.GetToolResult[ReadFilesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error reading multiple files")

		requireReadFilesResult(t, result, err, readFilesResultOpts{
			ExpectFiles:      2,
			ExpectedContents: []string{"Content 1", "Content 2"},
		})
	})

	t.Run("ReadDirectory_ShouldReturnAllFilesInDirectory", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReadFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("test-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf.AddFileFixtures(t, testutil.FileFixtureArgs{
			ContentFunc: func(ff *testutil.FileFixture) string {
				return fmt.Sprintf("Content of %s", ff.Name)
			},
			Permissions: 0644,
		}, "README.md", "main.go", "config.yaml")

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"paths":         []any{pf.Dir},
			"recursive":     true,
		})

		result, err := mcputil.GetToolResult[ReadFilesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error reading directory")

		requireReadFilesResult(t, result, err, readFilesResultOpts{
			MinFiles:         3, // At least the 3 files we created
			ExpectedContents: []string{"Content of README.md", "Content of main.go", "Content of config.yaml"},
		})
	})

	t.Run("ReadNonexistentFile_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReadFilesDirPrefix)
		defer tf.Cleanup()

		// Add a missing file that doesn't exist
		missingFile := tf.AddFileFixture("does-not-exist.txt", testutil.FileFixtureArgs{
			Missing: true,
		})
		tf.Setup(t)

		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"paths":         []any{missingFile.Filepath},
		})

		result, err := mcputil.GetToolResult[ReadFilesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should handle nonexistent file gracefully")

		requireReadFilesResult(t, result, err, readFilesResultOpts{
			ExpectPartialError: true,
			ExpectedErrorMsg:   "no such file",
		})
	})
}
