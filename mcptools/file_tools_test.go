package mcptools_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const FileDirPrefix = "file-tools-test"

type fileToolResultOpts struct {
	ExpectError        bool
	ExpectedErrorMsg   string
	ExpectPartialError bool
	ExpectFiles        int
	MinFiles           int
	ExpectedContent    string
	ExpectedContents   []string
	ExpectedSize       int64
	ExpectedFilePath   string
	ShouldCreateFile   bool
	ShouldUpdateFile   bool
	ShouldDeleteFile   string
}

func requireFileToolResult(t *testing.T, result *map[string]any, err error, opts fileToolResultOpts) {
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

	// Check for files in result
	if opts.ExpectFiles > 0 || opts.MinFiles > 0 || len(opts.ExpectedContents) > 0 {
		filesInterface, hasFiles := (*result)["files"]
		if hasFiles {
			files, ok := filesInterface.([]any)
			require.True(t, ok, "Files should be an array")

			if opts.ExpectFiles > 0 {
				assert.Len(t, files, opts.ExpectFiles, "Should have expected number of files")
			}
			if opts.MinFiles > 0 {
				assert.GreaterOrEqual(t, len(files), opts.MinFiles, "Should have at least minimum number of files")
			}

			// Check contents
			if opts.ExpectedContent != "" {
				found := false
				for _, fileInterface := range files {
					if file, ok := fileInterface.(map[string]any); ok {
						if content, hasContent := file["content"].(string); hasContent && content == opts.ExpectedContent {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Should find expected content: %s", opts.ExpectedContent)
			}

			if len(opts.ExpectedContents) > 0 {
				foundContents := make(map[string]bool)
				for _, fileInterface := range files {
					if file, ok := fileInterface.(map[string]any); ok {
						if content, hasContent := file["content"].(string); hasContent {
							foundContents[content] = true
						}
					}
				}
				for _, expectedContent := range opts.ExpectedContents {
					assert.True(t, foundContents[expectedContent], "Should find expected content: %s", expectedContent)
				}
			}
		}
	}

	// Check for partial errors in results
	if opts.ExpectPartialError {
		errorsInterface, hasErrors := (*result)["errors"]
		if hasErrors {
			errors, ok := errorsInterface.([]any)
			require.True(t, ok, "Errors should be an array")
			assert.Greater(t, len(errors), 0, "Should have at least one error")

			if opts.ExpectedErrorMsg != "" {
				found := false
				for _, errorInterface := range errors {
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

	if opts.ShouldDeleteFile != "" {
		_, err := os.Stat(opts.ShouldDeleteFile)
		assert.True(t, os.IsNotExist(err), "File should be deleted: %s", opts.ShouldDeleteFile)
	}
}

func TestReadFilesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("read_files")
	require.NotNil(t, tool, "read_files tool should be registered")

	t.Run("ReadSingleFile_ShouldReturnFileContentAndMetadata", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		// Create a test file with known content
		pf := tf.AddProjectFixture("test-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("test.txt", FileFixtureArgs{
			Content:     "Hello, World!",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"paths":         []any{testFile.Filepath},
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error reading single file",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ExpectFiles:     1,
			ExpectedContent: "Hello, World!",
			ExpectedSize:    13,
		})
	})

	t.Run("ReadMultipleFiles_ShouldReturnAllFileContents", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("test-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf1 := pf.AddFileFixture("file1.txt", FileFixtureArgs{
			Content:     "Content 1",
			Permissions: 0644,
		})
		pf2 := pf.AddFileFixture("file2.txt", FileFixtureArgs{
			Content:     "Content 2",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"paths": []any{
				pf1.Filepath,
				pf2.Filepath,
			},
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error reading multiple files",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ExpectFiles:      2,
			ExpectedContents: []string{"Content 1", "Content 2"},
		})
	})

	t.Run("ReadDirectory_ShouldReturnAllFilesInDirectory", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("test-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf.AddFileFixtures(t, FileFixtureArgs{
			ContentFunc: func(ff *FileFixture) string {
				return fmt.Sprintf("Content of %s", ff.Name)
			},
			Permissions: 0644,
		}, "README.md", "main.go", "config.yaml")

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"paths":         []any{pf.Dir},
			"recursive":     true,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error reading directory",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			MinFiles:         3, // At least the 3 files we created
			ExpectedContents: []string{"Content of README.md", "Content of main.go", "Content of config.yaml"},
		})
	})

	t.Run("ReadNonexistentFile_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		// Add a missing file that doesn't exist
		missingFile := tf.AddFileFixture("does-not-exist.txt", FileFixtureArgs{
			Missing: true,
		})
		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"paths":         []any{missingFile.Filepath},
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle nonexistent file gracefully",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ExpectPartialError: true,
			ExpectedErrorMsg:   "no such file",
		})
	})
}

func TestSearchFilesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("search_files")
	require.NotNil(t, tool, "search_files tool should be registered")

	t.Run("BasicSearch_ShouldReturnAllFiles", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("search-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf.AddFileFixtures(t, FileFixtureArgs{
			Permissions: 0644,
		}, "file1.txt", "file2.go", "README.md")

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          pf.Dir,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error searching files",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			MinFiles: 3, // At least our 3 files
		})
	})

	t.Run("SearchWithPattern_ShouldReturnMatchingFiles", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("pattern-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf.AddFileFixtures(t, FileFixtureArgs{
			Permissions: 0644,
		}, "test-file.txt", "other-file.go", "test-config.yaml")

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          pf.Dir,
			"pattern":       "test",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error searching with pattern",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ExpectFiles: 2, // Should find test-file.txt and test-config.yaml
		})
	})

	t.Run("SearchWithExtensions_ShouldReturnOnlyMatchingExtensions", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("ext-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf.AddFileFixtures(t, FileFixtureArgs{
			Permissions: 0644,
		}, "main.go", "utils.go", "config.yaml", "README.md")

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          pf.Dir,
			"extensions":    []any{".go"},
			"recursive":     true,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error searching with extensions",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ExpectFiles: 2, // Should find only main.go and utils.go
		})
	})
}

func TestCreateFileTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("create_file")
	require.NotNil(t, tool, "create_file tool should be registered")

	t.Run("CreateNewFile_ShouldCreateFileWithContent", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()
		// Add a pending file that will be created by the tool
		newFile := tf.AddFileFixture("new_file.txt", FileFixtureArgs{
			Pending: true,
		})
		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      newFile.Filepath,
			"new_content":   "This is new content",
			"create_dirs":   true,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error creating file",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ShouldCreateFile: true,
			ExpectedFilePath: newFile.Filepath,
			ExpectedContent:  "This is new content",
		})
	})

	t.Run("CreateFileWithDirectories_ShouldCreateParentDirectories", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		// Add a pending file in nested directories
		newFile := tf.AddFileFixture("new_dir/nested_dir/new_file.txt", FileFixtureArgs{
			Pending: true,
		})

		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      newFile.Filepath,
			"new_content":   "Content in nested directory",
			"create_dirs":   true,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error creating file with nested directories",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
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
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		// Add a pending file in missing directory (should fail without create_dirs)
		newFile := tf.AddFileFixture("missing_dir/new_file.txt", FileFixtureArgs{
			Pending: true,
		})
		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      newFile.Filepath,
			"new_content":   "This should fail",
			"create_dirs":   false,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle missing parent directory",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "no such file or directory",
		})
	})
}

func TestUpdateFileTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("update_file")
	require.NotNil(t, tool, "update_file tool should be registered")

	t.Run("UpdateExistingFile_ShouldReplaceFileContent", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("update-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("update_me.txt", FileFixtureArgs{
			Content:     "Original content",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      testFile.Filepath,
			"new_content":   "Updated content",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error updating file",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ShouldCreateFile: true, // File should still exist
			ExpectedFilePath: testFile.Filepath,
			ExpectedContent:  "Updated content",
		})
	})

	t.Run("UpdateNonexistentFile_ShouldReturnError", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()
		// Add a missing file for error testing
		nonexistentFile := tf.AddFileFixture("does-not-exist.txt", FileFixtureArgs{
			Missing: true,
		})
		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"filepath":      nonexistentFile.Filepath,
			"new_content":   "This should fail",
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle nonexistent file",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "no such file or directory",
		})
	})
}

func TestDeleteFilesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("delete_files")
	require.NotNil(t, tool, "delete_files tool should be registered")

	t.Run("DeleteExistingFile_ShouldRemoveFileFromDisk", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("delete-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		fileToDelete := pf.AddFileFixture("delete_me.txt", FileFixtureArgs{
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
			"session_token": tf.token,
			"path":          fileToDelete.Filepath,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error deleting file",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ShouldDeleteFile: fileToDelete.Filepath,
		})
	})

	t.Run("DeleteNonexistentFile_ShouldReturnError", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		// Add a missing file for error testing
		nonexistentFile := tf.AddFileFixture("does-not-exist.txt", FileFixtureArgs{
			Missing: true,
		})

		tf.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          nonexistentFile.Filepath,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle nonexistent file",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "no such file or directory",
		})
	})

	t.Run("DeleteDirectory_ShouldRemoveDirectoryAndContents", func(t *testing.T) {
		tf := NewTestFixture(FileDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("delete-dir-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})

		// Add a subdirectory structure to delete
		subDirFile := pf.AddFileFixture("subdir/file.txt", FileFixtureArgs{
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
			"session_token": tf.token,
			"path":          subDir,
			"recursive":     true,
		})

		result, err := getToolResult[map[string]any](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error deleting directory",
		)

		requireFileToolResult(t, result, err, fileToolResultOpts{
			ShouldDeleteFile: subDir,
		})
	})
}
