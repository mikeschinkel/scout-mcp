package mcptools_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const SearchFilesDirPrefix = "search-files-tool-test"

// Search files tool result type
type SearchFilesResult struct {
	SearchPath string `json:"search_path"`
	Results    []struct {
		Path  string `json:"path"`
		Name  string `json:"name"`
		Size  int64  `json:"size"`
		IsDir bool   `json:"is_directory"`
	} `json:"results"`
	Count       int      `json:"count"`
	Recursive   bool     `json:"recursive"`
	Pattern     string   `json:"pattern,omitempty"`
	NamePattern string   `json:"name_pattern,omitempty"`
	Extensions  []string `json:"extensions,omitempty"`
	FilesOnly   bool     `json:"files_only"`
	DirsOnly    bool     `json:"dirs_only"`
	MaxResults  int      `json:"max_results"`
}

type searchFilesResultOpts struct {
	ExpectError      bool
	ExpectedErrorMsg string
	ExpectFiles      int
	MinFiles         int
}

func requireSearchFilesResult(t *testing.T, result *SearchFilesResult, err error, opts searchFilesResultOpts) {
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
		assert.Len(t, result.Results, opts.ExpectFiles, "Should have expected number of files")
	}
	if opts.MinFiles > 0 {
		assert.GreaterOrEqual(t, len(result.Results), opts.MinFiles, "Should have at least minimum number of files")
	}

	// Verify count matches array length
	assert.Equal(t, len(result.Results), result.Count, "Count should match results array length")
}

func TestSearchFilesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("search_files")
	require.NotNil(t, tool, "search_files tool should be registered")

	t.Run("BasicSearch_ShouldReturnAllFiles", func(t *testing.T) {
		tf := testutil.NewTestFixture(SearchFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("search-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf.AddFileFixtures(t, testutil.FileFixtureArgs{
			Permissions: 0644,
		}, "file1.txt", "file2.go", "README.md")

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          pf.Dir,
		})

		result, err := mcputil.GetToolResult[SearchFilesResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error searching files")

		requireSearchFilesResult(t, result, err, searchFilesResultOpts{
			MinFiles: 3, // At least our 3 files
		})
	})

	t.Run("SearchWithPattern_ShouldReturnMatchingFiles", func(t *testing.T) {
		tf := testutil.NewTestFixture(SearchFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("pattern-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf.AddFileFixtures(t, testutil.FileFixtureArgs{
			Permissions: 0644,
		}, "test-file.txt", "other-file.go", "test-config.yaml")

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          pf.Dir,
			"pattern":       "test",
		})

		result, err := mcputil.GetToolResult[SearchFilesResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error searching with pattern")

		requireSearchFilesResult(t, result, err, searchFilesResultOpts{
			ExpectFiles: 2, // Should find test-file.txt and test-config.yaml
		})
	})

	t.Run("SearchWithExtensions_ShouldReturnOnlyMatchingExtensions", func(t *testing.T) {
		tf := testutil.NewTestFixture(SearchFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("ext-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf.AddFileFixtures(t, testutil.FileFixtureArgs{
			Permissions: 0644,
		}, "main.go", "utils.go", "config.yaml", "README.md")

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          pf.Dir,
			"extensions":    []any{".go"},
			"recursive":     true,
		})

		result, err := mcputil.GetToolResult[SearchFilesResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error searching with extensions")

		requireSearchFilesResult(t, result, err, searchFilesResultOpts{
			ExpectFiles: 2, // Should find only main.go and utils.go
		})
	})
}
