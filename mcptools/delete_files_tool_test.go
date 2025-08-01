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

const DeleteFilesDirPrefix = "delete-files-tool-test"

// Delete files tool result type
type DeleteFilesResult struct {
	Success     bool   `json:"success"`
	DeletedPath string `json:"deleted_path"`
	FileType    string `json:"file_type"`
	Message     string `json:"message"`
}

type deleteFilesResultOpts struct {
	ExpectError      bool
	ExpectedErrorMsg string
	ExpectedPath     string
	ShouldDeleteFile string
}

func requireDeleteFilesResult(t *testing.T, result *DeleteFilesResult, err error, opts deleteFilesResultOpts) {
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

	if opts.ExpectedPath != "" {
		assert.Equal(t, opts.ExpectedPath, result.DeletedPath, "Deleted path should match expected")
	}

	// Check file system side effects
	if opts.ShouldDeleteFile != "" {
		_, err := os.Stat(opts.ShouldDeleteFile)
		assert.True(t, os.IsNotExist(err), "File should be deleted: %s", opts.ShouldDeleteFile)
	}
}

func TestDeleteFilesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("delete_files")
	require.NotNil(t, tool, "delete_files tool should be registered")

	t.Run("DeleteExistingFile_ShouldRemoveFileFromDisk", func(t *testing.T) {
		tf := testutil.NewTestFixture(DeleteFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("delete-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		fileToDelete := pf.AddFileFixture("delete_me.txt", testutil.FileFixtureArgs{
			Content:     "This will be deleted",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		// Verify file exists before deletion
		_, err := os.Stat(fileToDelete.Filepath)
		require.NoError(t, err, "File should exist before deletion")

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          fileToDelete.Filepath,
		})

		result, err := getToolResult[DeleteFilesResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error deleting file",
		)

		requireDeleteFilesResult(t, result, err, deleteFilesResultOpts{
			ExpectedPath:     fileToDelete.Filepath,
			ShouldDeleteFile: fileToDelete.Filepath,
		})
	})

	t.Run("DeleteNonexistentFile_ShouldReturnError", func(t *testing.T) {
		tf := testutil.NewTestFixture(DeleteFilesDirPrefix)
		defer tf.Cleanup()

		// Add a missing file for error testing
		nonexistentFile := tf.AddFileFixture("does-not-exist.txt", testutil.FileFixtureArgs{
			Missing: true,
		})

		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          nonexistentFile.Filepath,
		})

		result, err := getToolResult[DeleteFilesResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle nonexistent file",
		)

		requireDeleteFilesResult(t, result, err, deleteFilesResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "no such file or directory",
		})
	})

	t.Run("DeleteDirectory_ShouldRemoveDirectoryAndContents", func(t *testing.T) {
		tf := testutil.NewTestFixture(DeleteFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("delete-dir-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})

		// Add a subdirectory structure to delete
		subDirFile := pf.AddFileFixture("subdir/file.txt", testutil.FileFixtureArgs{
			Content:     "content",
			Permissions: 0644,
		})

		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		// We want to delete the parent directory, so get the directory path
		subDir := filepath.Dir(subDirFile.Filepath)

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          subDir,
			"recursive":     true,
		})

		result, err := getToolResult[DeleteFilesResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error deleting directory",
		)

		requireDeleteFilesResult(t, result, err, deleteFilesResultOpts{
			ExpectedPath:     subDir,
			ShouldDeleteFile: subDir,
		})
	})
}
