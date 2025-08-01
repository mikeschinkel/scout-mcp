package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStartSessionDirect tests start_session using direct server access
func TestStartSessionDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("TokenGeneration", func(t *testing.T) {
		// start_session was already called during env setup
		token := env.GetSessionToken()
		assert.NotEmpty(t, token, "Session token should not be empty")
		assert.Greater(t, len(token), 10, "Session token should be reasonably long")
	})

	t.Run("SessionResponseStructure", func(t *testing.T) {
		// Call start_session again to check response structure
		result := env.CallTool(t, "start_session", map[string]interface{}{})

		var response map[string]interface{}
		ParseJSONResult(t, result, &response)

		// Verify required fields exist
		assert.Contains(t, response, "session_token", "Response should contain session_token")
		assert.Contains(t, response, "token_expires_at", "Response should contain token_expires_at")
		assert.Contains(t, response, "quick_start", "Response should contain quick_start")
		assert.Contains(t, response, "server_config", "Response should contain server_config")
		assert.Contains(t, response, "instructions", "Response should contain instructions")
		assert.Contains(t, response, "message", "Response should contain message")

		// Verify current_project field exists (may be null for test environment)
		_, hasCurrentProject := response["current_project"]
		assert.True(t, hasCurrentProject, "Response should contain current_project field")
	})

	t.Run("SessionWithCurrentProject", func(t *testing.T) {
		// Create a fresh test environment with predictable project setup
		env := NewDirectServerTestEnv(t)
		defer env.Cleanup()

		// Clean up the default test files first
		testDir := env.GetTestDir()
		entries, _ := os.ReadDir(testDir)
		for _, entry := range entries {
			entryPath := filepath.Join(testDir, entry.Name())
			if entry.IsDir() {
				must(os.RemoveAll(entryPath))
			} else {
				must(os.Remove(entryPath))
			}
		}

		// Create multiple projects with predictable times
		// Project 1: Old project (25 hours ago) - should NOT be current
		oldProjectDir := filepath.Join(testDir, "old-project")
		createTestProjectWithTime(t, oldProjectDir, time.Now().Add(-25*time.Hour))

		// Project 2: Current project (1 hour ago) - should BE current
		currentProjectDir := filepath.Join(testDir, "current-project")
		createTestProjectWithTime(t, currentProjectDir, time.Now().Add(-1*time.Hour))

		// Call start_session and verify current project detection
		result := env.CallTool(t, "start_session", map[string]interface{}{})

		var response map[string]interface{}
		ParseJSONResult(t, result, &response)

		// Verify current_project is detected
		assert.Contains(t, response, "current_project", "Response should contain current_project")

		currentProjectData, ok := response["current_project"].(map[string]interface{})
		require.True(t, ok, "current_project should be an object")
		require.NotNil(t, currentProjectData, "current_project should not be null")

		// Should have a clear winner (24+ hour difference)
		assert.Contains(t, currentProjectData, "current_project", "Should identify a current project")
		assert.Contains(t, currentProjectData, "recent_projects", "Should include recent projects list")
		assert.Contains(t, currentProjectData, "summary", "Should include summary")

		currentProj, ok := currentProjectData["current_project"].(map[string]interface{})
		require.True(t, ok, "current_project should contain a project object")

		// Verify it detected the correct project
		assert.Equal(t, "current-project", currentProj["name"], "Should detect current-project as current")
		assert.Contains(t, currentProj["path"].(string), "current-project", "Path should contain current-project")

		// Should not require choice (clear 24+ hour winner)
		requiresChoice, _ := currentProjectData["requires_choice"].(bool)
		assert.False(t, requiresChoice, "Should not require choice with clear winner")
	})

	t.Run("SessionWithMultipleRecentProjects", func(t *testing.T) {
		// Create a fresh test environment
		env := NewDirectServerTestEnv(t)
		defer env.Cleanup()

		// Clean up the default test files first
		testDir := env.GetTestDir()
		entries, _ := os.ReadDir(testDir)
		for _, entry := range entries {
			entryPath := filepath.Join(testDir, entry.Name())
			if entry.IsDir() {
				must(os.RemoveAll(entryPath))
			} else {
				must(os.Remove(entryPath))
			}
		}

		// Create multiple recent projects (within 24 hours)
		project1Dir := filepath.Join(testDir, "recent-project-1")
		createTestProjectWithTime(t, project1Dir, time.Now().Add(-2*time.Hour))

		project2Dir := filepath.Join(testDir, "recent-project-2")
		createTestProjectWithTime(t, project2Dir, time.Now().Add(-3*time.Hour))

		// Call start_session
		result := env.CallTool(t, "start_session", map[string]interface{}{})

		var response map[string]interface{}
		ParseJSONResult(t, result, &response)

		// Verify current_project shows multiple recent projects
		currentProjectData, ok := response["current_project"].(map[string]interface{})
		require.True(t, ok, "current_project should be an object")
		require.NotNil(t, currentProjectData, "current_project should not be null")

		// Should require choice (multiple recent projects)
		requiresChoice, ok := currentProjectData["requires_choice"].(bool)
		assert.True(t, ok && requiresChoice, "Should require choice with multiple recent projects")

		assert.Contains(t, currentProjectData, "choice_message", "Should have choice message")
		assert.Contains(t, currentProjectData, "recent_projects", "Should have recent projects")

		recentProjects, ok := currentProjectData["recent_projects"].([]interface{})
		require.True(t, ok, "recent_projects should be an array")
		assert.GreaterOrEqual(t, len(recentProjects), 2, "Should have at least 2 recent projects")
	})
}

// TestGetConfigDirect tests get_config using direct server access
func TestGetConfigDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	result := env.CallTool(t, "get_config", map[string]interface{}{})

	// Parse the config response
	type ConfigInfo struct {
		ServerName     string   `json:"server_name"`
		AllowedPaths   []string `json:"allowed_paths"`
		AllowedOrigins []string `json:"allowed_origins"`
		PathCount      int      `json:"path_count"`
		ConfigFilePath string   `json:"config_file_path"`
		HomeDirectory  string   `json:"home_directory"`
		ServerPort     string   `json:"server_port"`
		Summary        string   `json:"summary"`
	}

	var config ConfigInfo
	ParseJSONResult(t, result, &config)

	// Verify config structure
	assert.NotEmpty(t, config.ServerName, "Server name should not be empty")
	assert.Greater(t, config.PathCount, 0, "Should have at least one allowed path")
	assert.Contains(t, config.AllowedPaths, env.GetTestDir(), "Test directory should be in allowed paths")
	assert.NotEmpty(t, config.HomeDirectory, "Home directory should not be empty")
	assert.NotEmpty(t, config.Summary, "Summary should not be empty")
}

// TestSearchFilesDirect tests search_files using direct server access
func TestSearchFilesDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("BasicSearch", func(t *testing.T) {
		result := env.CallTool(t, "search_files", map[string]interface{}{
			"path": env.GetTestDir(),
		})

		type SearchResponse struct {
			SearchPath string `json:"search_path"`
			Results    []struct {
				Path  string `json:"path"`
				Name  string `json:"name"`
				IsDir bool   `json:"is_directory"`
			} `json:"results"`
			Count int `json:"count"`
		}

		var searchResp SearchResponse
		ParseJSONResult(t, result, &searchResp)

		assert.Equal(t, env.GetTestDir(), searchResp.SearchPath, "Search path should match")
		assert.Greater(t, searchResp.Count, 0, "Should find some files")
		assert.Greater(t, len(searchResp.Results), 0, "Should have results")
	})

	t.Run("PatternSearch", func(t *testing.T) {
		result := env.CallTool(t, "search_files", map[string]interface{}{
			"path":    env.GetTestDir(),
			"pattern": "README",
		})

		type SearchResponse struct {
			Results []struct {
				Path string `json:"path"`
				Name string `json:"name"`
			} `json:"results"`
			Count int `json:"count"`
		}

		var searchResp SearchResponse
		ParseJSONResult(t, result, &searchResp)

		assert.Greater(t, searchResp.Count, 0, "Should find README file")
		found := false
		for _, result := range searchResp.Results {
			if result.Name == "README.md" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find README.md file")
	})

	t.Run("ExtensionFilter", func(t *testing.T) {
		result := env.CallTool(t, "search_files", map[string]interface{}{
			"path":       env.GetTestDir(),
			"extensions": []string{".go"},
			"recursive":  true,
		})

		type SearchResponse struct {
			Results []struct {
				Path string `json:"path"`
				Name string `json:"name"`
			} `json:"results"`
			Count int `json:"count"`
		}

		var searchResp SearchResponse
		ParseJSONResult(t, result, &searchResp)

		assert.Greater(t, searchResp.Count, 0, "Should find Go files")
		for _, result := range searchResp.Results {
			assert.True(t, filepath.Ext(result.Name) == ".go", "All results should be .go files")
		}
	})
}

// TestReadFilesDirect tests read_files using direct server access
func TestReadFilesDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("ReadSingleFile", func(t *testing.T) {
		readmePath := filepath.Join(env.GetTestDir(), "README.md")

		result := env.CallTool(t, "read_files", map[string]interface{}{
			"paths": []string{readmePath},
		})

		type ReadFilesResponse struct {
			Files []struct {
				Path    string `json:"path"`
				Name    string `json:"name"`
				Content string `json:"content"`
				Size    int64  `json:"size"`
			} `json:"files"`
			TotalFiles int `json:"total_files"`
		}

		var readResp ReadFilesResponse
		ParseJSONResult(t, result, &readResp)

		assert.Equal(t, 1, readResp.TotalFiles, "Should read exactly one file")
		assert.Len(t, readResp.Files, 1, "Should have one file result")
		assert.Contains(t, readResp.Files[0].Content, "Test Project", "File content should contain expected text")
	})

	t.Run("ReadMultipleFiles", func(t *testing.T) {
		testFile1 := filepath.Join(env.GetTestDir(), "README.md")
		testFile2 := filepath.Join(env.GetTestDir(), "src/main.go")

		result := env.CallTool(t, "read_files", map[string]interface{}{
			"paths": []string{testFile1, testFile2},
		})

		type ReadFilesResponse struct {
			Files []struct {
				Path    string `json:"path"`
				Name    string `json:"name"`
				Content string `json:"content"`
			} `json:"files"`
			TotalFiles int `json:"total_files"`
		}

		var readResp ReadFilesResponse
		ParseJSONResult(t, result, &readResp)

		assert.Equal(t, 2, readResp.TotalFiles, "Should read exactly two files")
		assert.Len(t, readResp.Files, 2, "Should have two file results")
	})

	t.Run("ReadDirectory", func(t *testing.T) {
		result := env.CallTool(t, "read_files", map[string]interface{}{
			"paths":     []string{env.GetTestDir()},
			"recursive": true,
			"max_files": 10,
		})

		type ReadFilesResponse struct {
			Files []struct {
				Path string `json:"path"`
				Name string `json:"name"`
			} `json:"files"`
			TotalFiles int `json:"total_files"`
		}

		var readResp ReadFilesResponse
		ParseJSONResult(t, result, &readResp)

		assert.Greater(t, readResp.TotalFiles, 0, "Should read some files from directory")
		assert.Greater(t, len(readResp.Files), 0, "Should have file results")
	})
}

// TestFileOperationsDirect tests file creation and manipulation using direct server access
func TestFileOperationsDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("CreateFile", func(t *testing.T) {
		testFilePath := filepath.Join(env.GetTestDir(), "created_file.txt")
		testContent := "This is a test file created by direct integration test"

		// Remove file if it exists (ignore error if file doesn't exist)
		testutil.MaybeRemove(t, testFilePath)

		result := env.CallTool(t, "create_file", map[string]interface{}{
			"filepath":    testFilePath,
			"new_content": testContent,
			"create_dirs": true,
		})

		// Verify the result
		assert.NotNil(t, result, "create_file should return result")

		// Verify file was created with correct content
		createdContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Should be able to read created file")
		assert.Equal(t, testContent, string(createdContent), "Created file content should match")
	})

	t.Run("UpdateFile", func(t *testing.T) {
		testFilePath := filepath.Join(env.GetTestDir(), "README.md")
		originalContent := "# Test Project\nThis is a test project for Scout MCP testing.\n"
		updatedContent := "# Updated Test Project\nThis file has been updated by the integration test.\n"

		// Verify original content
		content, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Should be able to read original file")
		assert.Equal(t, originalContent, string(content), "Original content should match")

		result := env.CallTool(t, "update_file", map[string]interface{}{
			"filepath":    testFilePath,
			"new_content": updatedContent,
		})

		// Verify the result
		assert.NotNil(t, result, "update_file should return result")

		// Verify file was updated
		finalContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Should be able to read updated file")
		assert.Equal(t, updatedContent, string(finalContent), "Updated file content should match")
	})

	t.Run("DeleteFile", func(t *testing.T) {
		// Create a file to delete
		fileToDelete := filepath.Join(env.GetTestDir(), "to_delete.txt")
		err := os.WriteFile(fileToDelete, []byte("delete me"), 0644)
		require.NoError(t, err, "Should be able to create file to delete")

		result := env.CallTool(t, "delete_files", map[string]interface{}{
			"path": fileToDelete,
		})

		// Verify the result
		assert.NotNil(t, result, "delete_files should return result")

		// Verify file was deleted
		_, err = os.Stat(fileToDelete)
		assert.True(t, os.IsNotExist(err), "File should be deleted")
	})
}

// createTestProjectWithTime creates a project directory with .git and enough files, then sets specific modification times
func createTestProjectWithTime(t *testing.T, projectDir string, modTime time.Time) {
	// Create project directory
	err := os.Mkdir(projectDir, 0755)
	require.NoError(t, err, "Failed to create project directory %s", projectDir)

	// Create .git directory to make it a valid project
	gitDir := filepath.Join(projectDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	require.NoError(t, err, "Failed to create .git directory for %s", projectDir)

	// Create enough files (5+) to meet threshold
	testFiles := []string{"README.md", "main.go", "config.json", "package.json", "Makefile", "LICENSE"}
	for _, fileName := range testFiles {
		filePath := filepath.Join(projectDir, fileName)
		err = os.WriteFile(filePath, []byte("test content for "+filepath.Base(projectDir)), 0644)
		require.NoError(t, err, "Failed to create test file %s in %s", fileName, projectDir)

		// Set specific modification time for the file
		err = os.Chtimes(filePath, modTime, modTime)
		require.NoError(t, err, "Failed to set modification time for %s", filePath)
	}
}
