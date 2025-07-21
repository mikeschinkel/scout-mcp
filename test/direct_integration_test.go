package test

import (
	"os"
	"path/filepath"
	"testing"

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
		assert.Contains(t, response, "tool_help", "Response should contain tool_help")
		assert.Contains(t, response, "server_config", "Response should contain server_config")
		assert.Contains(t, response, "instructions", "Response should contain instructions")
		assert.Contains(t, response, "message", "Response should contain message")
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

		// Remove file if it exists
		testutil.Must(t, os.Remove(testFilePath))

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
