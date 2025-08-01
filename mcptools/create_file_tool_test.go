package mcptools_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const CreateFileDirPrefix = "create-file-tool-test"

// Create file tool result type
type CreateFileResult struct {
	Success  bool   `json:"success"`
	FilePath string `json:"file_path"`
	Size     int64  `json:"size"`
	Message  string `json:"message"`
}

type createFileResultOpts struct {
	ExpectError      bool
	ExpectedErrorMsg string
	ExpectedFilePath string
	ShouldCreateFile bool
	ExpectedContent  string
}

func requireCreateFileResult(t *testing.T, result *CreateFileResult, err error, opts createFileResultOpts) {
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
	if opts.ShouldCreateFile && opts.ExpectedFilePath != "" {
		_, err := os.Stat(opts.ExpectedFilePath)
		assert.NoError(t, err, "File should exist on disk: %s", opts.ExpectedFilePath)

		if opts.ExpectedContent != "" {
			content, readErr := os.ReadFile(opts.ExpectedFilePath)
			require.NoError(t, readErr, "Should be able to read created file")
			assert.Equal(t, opts.ExpectedContent, string(content), "File content should match expected")
		}
	}
}

func TestCreateFileTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("create_file")
	require.NotNil(t, tool, "create_file tool should be registered")

	t.Run("CreateNewFile_ShouldCreateFileWithContent", func(t *testing.T) {
		tf := testutil.NewTestFixture(CreateFileDirPrefix)
		defer tf.Cleanup()
		// Add a pending file that will be created by the tool
		newFile := tf.AddFileFixture("new_file.txt", testutil.FileFixtureArgs{
			Pending: true,
		})
		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"filepath":      newFile.Filepath,
			"new_content":   "This is new content",
			"create_dirs":   true,
		})

		result, err := getToolResult[CreateFileResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error creating file",
		)

		requireCreateFileResult(t, result, err, createFileResultOpts{
			ShouldCreateFile: true,
			ExpectedFilePath: newFile.Filepath,
			ExpectedContent:  "This is new content",
		})
	})

	t.Run("CreateFileWithDirectories_ShouldCreateParentDirectories", func(t *testing.T) {
		tf := testutil.NewTestFixture(CreateFileDirPrefix)
		defer tf.Cleanup()

		// Add a pending file in nested directories
		newFile := tf.AddFileFixture("new_dir/nested_dir/new_file.txt", testutil.FileFixtureArgs{
			Pending: true,
		})

		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"filepath":      newFile.Filepath,
			"new_content":   "Content in nested directory",
			"create_dirs":   true,
		})

		result, err := getToolResult[CreateFileResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error creating file with nested directories",
		)

		requireCreateFileResult(t, result, err, createFileResultOpts{
			ShouldCreateFile: true,
			ExpectedFilePath: newFile.Filepath,
			ExpectedContent:  "Content in nested directory",
		})

		// Verify parent directories were created
		parentDir := filepath.Dir(newFile.Filepath)
		info, statErr := os.Stat(parentDir)
		require.NoError(t, statErr, "Parent directory should exist")
		assert.True(t, info.IsDir(), "Parent should be a directory")
	})

	t.Run("CreateFileWithoutCreateDirs_ShouldFailIfParentMissing", func(t *testing.T) {
		tf := testutil.NewTestFixture(CreateFileDirPrefix)
		defer tf.Cleanup()

		// Add a pending file in missing directory (should fail without create_dirs)
		newFile := tf.AddFileFixture("missing_dir/new_file.txt", testutil.FileFixtureArgs{
			Pending: true,
		})
		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"filepath":      newFile.Filepath,
			"new_content":   "This should fail",
			"create_dirs":   false,
		})

		result, err := getToolResult[CreateFileResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle missing parent directory",
		)

		requireCreateFileResult(t, result, err, createFileResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "no such file or directory",
		})
	})
}
