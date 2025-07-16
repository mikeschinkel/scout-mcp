package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig represents the structure we expect from get_config
type TestConfig struct {
	ServerName     string   `json:"server_name"`
	AllowedPaths   []string `json:"allowed_paths"`
	AllowedOrigins []string `json:"allowed_origins"`
	PathCount      int      `json:"path_count"`
	ConfigFilePath string   `json:"config_file_path"`
	HomeDirectory  string   `json:"home_directory"`
	ServerPort     string   `json:"server_port"`
	Summary        string   `json:"summary"`
}

// TestFileResult represents file search results
type TestFileResult struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	IsDir    bool   `json:"is_directory"`
}

// TestSearchResponse represents search_files response
type TestSearchResponse struct {
	SearchPath  string           `json:"search_path"`
	Results     []TestFileResult `json:"results"`
	Count       int              `json:"count"`
	Recursive   bool             `json:"recursive"`
	Pattern     string           `json:"pattern"`
	NamePattern string           `json:"name_pattern"`
	Extensions  []string         `json:"extensions"`
	FilesOnly   bool             `json:"files_only"`
	DirsOnly    bool             `json:"dirs_only"`
	MaxResults  int              `json:"max_results"`
	Truncated   bool             `json:"truncated"`
}

// TestReadFilesResult represents the structure returned by read_files
type TestReadFilesResult struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Size    int64  `json:"size"`
	Error   string `json:"error,omitempty"`
}

// TestReadFilesResponse represents read_files response
type TestReadFilesResponse struct {
	Files      []TestReadFilesResult `json:"files"`
	TotalFiles int                   `json:"total_files"`
	TotalSize  int64                 `json:"total_size"`
	Errors     []string              `json:"errors"`
	Paths      []string              `json:"paths"`
	Extensions []string              `json:"extensions"`
	Recursive  bool                  `json:"recursive"`
	Pattern    string                `json:"pattern"`
	MaxFiles   int                   `json:"max_files"`
	Truncated  bool                  `json:"truncated"`
}

// TestSessionResponse represents start_session response
type TestSessionResponse struct {
	SessionToken   string                 `json:"session_token"`
	TokenExpiresAt string                 `json:"token_expires_at"`
	ToolHelp       string                 `json:"tool_help"`
	ServerConfig   map[string]interface{} `json:"server_config"`
	Instructions   map[string]interface{} `json:"instructions"`
	Message        string                 `json:"message"`
}

func TestListTools(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()

	resp, err := client.ListTools(ctx)
	require.NoError(t, err, "Failed to list tools")
	require.Nil(t, resp.Error, "ListTools returned error: %v", resp.Error)

	// Parse the tools list
	var result map[string]interface{}
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to parse tools list response")

	tools, ok := result["tools"].([]interface{})
	require.True(t, ok, "Tools should be an array")
	require.Greater(t, len(tools), 0, "Should have at least one tool")

	// Check for expected tools including session management and new read_files
	expectedTools := []string{
		"start_session",
		"get_config",
		"search_files",
		"read_files", // Updated from read_file
		"create_file",
		"update_file",
		"delete_files",
		"tool_help",
	}

	foundTools := make(map[string]bool)
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		require.True(t, ok, "Tool should be an object")

		name, ok := toolMap["name"].(string)
		require.True(t, ok, "Tool should have a name")

		foundTools[name] = true
	}

	for _, expectedTool := range expectedTools {
		assert.True(t, foundTools[expectedTool], "Expected tool %s should be available", expectedTool)
	}

	// Ensure we have a reasonable number of tools (should be 19 total)
	assert.GreaterOrEqual(t, len(foundTools), 15, "Should have at least 15 tools")
}

func TestStartSession(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()

	resp, err := client.CallTool(ctx, "start_session", map[string]interface{}{})
	require.NoError(t, err, "Failed to call start_session")
	require.Nil(t, resp.Error, "start_session returned error: %v", resp.Error)

	var sessionResp TestSessionResponse
	parseToolResponse(t, resp, &sessionResp)

	// Validate session response structure
	assert.NotEmpty(t, sessionResp.SessionToken, "Session token should not be empty")
	assert.NotEmpty(t, sessionResp.TokenExpiresAt, "Token expiration should not be empty")
	assert.NotEmpty(t, sessionResp.ToolHelp, "Tool help should not be empty")
	assert.NotEmpty(t, sessionResp.ServerConfig, "Server config should not be empty")
	assert.NotEmpty(t, sessionResp.Instructions, "Instructions should not be empty")
	assert.Contains(t, sessionResp.Message, "Session Started Successfully", "Should contain success message")
}

func TestGetConfigWithSession(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()

	// First get a session token
	sessionResp, err := client.CallTool(ctx, "start_session", map[string]interface{}{})
	require.NoError(t, err, "Failed to call start_session")
	require.Nil(t, sessionResp.Error, "start_session returned error: %v", sessionResp.Error)

	var session TestSessionResponse
	parseToolResponse(t, sessionResp, &session)

	// Now call get_config with session token
	resp, err := client.CallTool(ctx, "get_config", map[string]interface{}{
		"session_token": session.SessionToken,
	})
	require.NoError(t, err, "Failed to call get_config")
	require.Nil(t, resp.Error, "get_config returned error: %v", resp.Error)

	// Parse the config response
	var result map[string]interface{}
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to parse get_config response")

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Result should have content array")
	require.Greater(t, len(content), 0, "Content should not be empty")

	// Parse the actual config from the first content item
	contentItem, ok := content[0].(map[string]interface{})
	require.True(t, ok, "Content item should be an object")

	text, ok := contentItem["text"].(string)
	require.True(t, ok, "Content item should have text")

	var config TestConfig
	err = json.Unmarshal([]byte(text), &config)
	require.NoError(t, err, "Failed to parse config JSON")

	// Validate config structure
	assert.NotEmpty(t, config.ServerName, "Server name should not be empty")
	assert.Greater(t, len(config.AllowedPaths), 0, "Should have at least one allowed path")
	assert.Equal(t, len(config.AllowedPaths), config.PathCount, "Path count should match allowed paths length")
	assert.NotEmpty(t, config.ServerPort, "Server port should not be empty")
	assert.Contains(t, config.AllowedOrigins, "https://claude.ai", "Should allow claude.ai")
}

func TestSearchFilesWithSession(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()
	testDir := GetTestDir()

	// Get session token first
	sessionResp, err := client.CallTool(ctx, "start_session", map[string]interface{}{})
	require.NoError(t, err, "Failed to call start_session")
	require.Nil(t, sessionResp.Error, "start_session returned error: %v", sessionResp.Error)

	var session TestSessionResponse
	parseToolResponse(t, sessionResp, &session)

	t.Run("BasicSearch", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "search_files", map[string]interface{}{
			"session_token": session.SessionToken,
			"path":          testDir,
			"recursive":     true,
		})
		require.NoError(t, err, "Failed to call search_files")
		require.Nil(t, resp.Error, "search_files returned error: %v", resp.Error)

		var searchResp TestSearchResponse
		parseToolResponse(t, resp, &searchResp)

		assert.Equal(t, testDir, searchResp.SearchPath, "Search path should match")
		assert.Greater(t, searchResp.Count, 0, "Should find at least one file")
		assert.Equal(t, len(searchResp.Results), searchResp.Count, "Count should match results length")
		assert.True(t, searchResp.Recursive, "Should be recursive")
	})

	t.Run("PatternSearch", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "search_files", map[string]interface{}{
			"session_token": session.SessionToken,
			"path":          testDir,
			"pattern":       "main",
			"recursive":     true,
		})
		require.NoError(t, err, "Failed to call search_files")
		require.Nil(t, resp.Error, "search_files returned error: %v", resp.Error)

		var searchResp TestSearchResponse
		parseToolResponse(t, resp, &searchResp)

		assert.Greater(t, searchResp.Count, 0, "Should find files with 'main' in name")
		for _, result := range searchResp.Results {
			assert.Contains(t, result.Name, "main", "All results should contain 'main'")
		}
	})

	t.Run("ExtensionFilter", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "search_files", map[string]interface{}{
			"session_token": session.SessionToken,
			"path":          testDir,
			"extensions":    []string{".go"},
			"recursive":     true,
		})
		require.NoError(t, err, "Failed to call search_files")
		require.Nil(t, resp.Error, "search_files returned error: %v", resp.Error)

		var searchResp TestSearchResponse
		parseToolResponse(t, resp, &searchResp)

		for _, result := range searchResp.Results {
			if !result.IsDir {
				assert.True(t,
					strings.HasSuffix(result.Name, ".go"),
					"File %s should have .go extension", result.Name)
			}
		}
	})
}

func TestReadFilesWithSession(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()
	testDir := GetTestDir()

	// Get session token first
	sessionResp, err := client.CallTool(ctx, "start_session", map[string]interface{}{})
	require.NoError(t, err, "Failed to call start_session")
	require.Nil(t, sessionResp.Error, "start_session returned error: %v", sessionResp.Error)

	var session TestSessionResponse
	parseToolResponse(t, sessionResp, &session)

	t.Run("ReadSingleFile", func(t *testing.T) {
		testFilePath := filepath.Join(testDir, "README.md")

		resp, err := client.CallTool(ctx, "read_files", map[string]interface{}{
			"session_token": session.SessionToken,
			"paths":         []string{testFilePath},
		})
		require.NoError(t, err, "Failed to call read_files")
		require.Nil(t, resp.Error, "read_files returned error: %v", resp.Error)

		var readResp TestReadFilesResponse
		parseToolResponse(t, resp, &readResp)

		assert.Equal(t, 1, readResp.TotalFiles, "Should read exactly one file")
		assert.Greater(t, len(readResp.Files), 0, "Should have file results")
		assert.Contains(t, readResp.Files[0].Content, "Test Project", "File content should contain expected text")
		assert.Contains(t, readResp.Files[0].Content, "Scout MCP", "File content should contain Scout MCP reference")
	})

	t.Run("ReadMultipleFiles", func(t *testing.T) {
		testFile1 := filepath.Join(testDir, "README.md")
		testFile2 := filepath.Join(testDir, "main.go")

		resp, err := client.CallTool(ctx, "read_files", map[string]interface{}{
			"session_token": session.SessionToken,
			"paths":         []string{testFile1, testFile2},
		})
		require.NoError(t, err, "Failed to call read_files")
		require.Nil(t, resp.Error, "read_files returned error: %v", resp.Error)

		var readResp TestReadFilesResponse
		parseToolResponse(t, resp, &readResp)

		assert.Equal(t, 2, readResp.TotalFiles, "Should read exactly two files")
		assert.Equal(t, len(readResp.Files), 2, "Should have two file results")
	})

	t.Run("ReadDirectory", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "read_files", map[string]interface{}{
			"session_token": session.SessionToken,
			"paths":         []string{testDir},
			"extensions":    []string{".md"},
			"recursive":     false,
		})
		require.NoError(t, err, "Failed to call read_files")
		require.Nil(t, resp.Error, "read_files returned error: %v", resp.Error)

		var readResp TestReadFilesResponse
		parseToolResponse(t, resp, &readResp)

		assert.Greater(t, readResp.TotalFiles, 0, "Should read at least one .md file")
		for _, file := range readResp.Files {
			assert.True(t, strings.HasSuffix(file.Name, ".md"), "All files should be .md files")
		}
	})
}

func TestCreateFileWithSession(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()
	testDir := GetTestDir()

	// Get session token first
	sessionResp, err := client.CallTool(ctx, "start_session", map[string]interface{}{})
	require.NoError(t, err, "Failed to call start_session")
	require.Nil(t, sessionResp.Error, "start_session returned error: %v", sessionResp.Error)

	var session TestSessionResponse
	parseToolResponse(t, sessionResp, &session)

	testContent := "Created by integration test with session"
	testFilePath := filepath.Join(testDir, "test_created_with_session.txt")

	// Ensure file doesn't exist
	err = removeFile(testFilePath)
	if err != nil {
		t.Errorf("Failed to remove existing test file: %v", err)
	}

	// Test creating the file with session token
	resp, err := client.CallTool(ctx, "create_file", map[string]interface{}{
		"session_token": session.SessionToken,
		"path":          testFilePath,
		"content":       testContent,
	})
	require.NoError(t, err, "Failed to call create_file")
	require.Nil(t, resp.Error, "create_file returned error: %v", resp.Error)

	// Verify file was created with correct content
	createdContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read created file")
	assert.Equal(t, testContent, string(createdContent), "Created file content should match")
}

func TestUpdateFileWithSession(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()
	testDir := GetTestDir()

	// Get session token first
	sessionResp, err := client.CallTool(ctx, "start_session", map[string]interface{}{})
	require.NoError(t, err, "Failed to call start_session")
	require.Nil(t, sessionResp.Error, "start_session returned error: %v", sessionResp.Error)

	var session TestSessionResponse
	parseToolResponse(t, sessionResp, &session)

	originalContent := "Original content for update test with session"
	updatedContent := "Updated content for update test with session"
	testFilePath := filepath.Join(testDir, "test_update_with_session.txt")

	// Create initial file
	err = os.WriteFile(testFilePath, []byte(originalContent), 0644)
	require.NoError(t, err, "Failed to create initial file")

	// Test updating the file with session token
	resp, err := client.CallTool(ctx, "update_file", map[string]interface{}{
		"session_token": session.SessionToken,
		"path":          testFilePath,
		"content":       updatedContent,
	})
	require.NoError(t, err, "Failed to call update_file")
	require.Nil(t, resp.Error, "update_file returned error: %v", resp.Error)

	// Verify file was updated
	finalContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")
	assert.Equal(t, updatedContent, string(finalContent), "Updated file content should match")
}

func TestDeleteFileWithSession(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()
	testDir := GetTestDir()

	// Get session token first
	sessionResp, err := client.CallTool(ctx, "start_session", map[string]interface{}{})
	require.NoError(t, err, "Failed to call start_session")
	require.Nil(t, sessionResp.Error, "start_session returned error: %v", sessionResp.Error)

	var session TestSessionResponse
	parseToolResponse(t, sessionResp, &session)

	testFilePath := filepath.Join(testDir, "test_delete_with_session.txt")

	// Create file to delete
	err = os.WriteFile(testFilePath, []byte("Delete me with session"), 0644)
	require.NoError(t, err, "Failed to create file to delete")

	// Verify file exists before deletion
	_, err = os.Stat(testFilePath)
	require.NoError(t, err, "File should exist before deletion")

	// Test deleting the file with session token
	resp, err := client.CallTool(ctx, "delete_files", map[string]interface{}{
		"session_token": session.SessionToken,
		"path":          testFilePath,
	})
	require.NoError(t, err, "Failed to call delete_files")
	require.Nil(t, resp.Error, "delete_files returned error: %v", resp.Error)

	// Verify file was deleted
	_, err = os.Stat(testFilePath)
	assert.True(t, os.IsNotExist(err), "File should be deleted")
}

func TestSessionTokenValidation(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()
	testDir := GetTestDir()

	t.Run("InvalidSessionToken", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "get_config", map[string]interface{}{
			"session_token": "invalid_token",
		})
		require.NoError(t, err, "Failed to call get_config")
		assert.NotNil(t, resp.Error, "get_config should return error for invalid session token")
		assert.Contains(t, resp.Error.Message, "invalid or expired session token", "Error should mention invalid session token")
	})

	t.Run("MissingSessionToken", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "search_files", map[string]interface{}{
			"path": testDir,
		})
		require.NoError(t, err, "Failed to call search_files")
		assert.NotNil(t, resp.Error, "search_files should return error for missing session token")
		assert.Contains(t, resp.Error.Message, "invalid or expired session token", "Error should mention missing session token")
	})
}

func TestErrorHandling(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()

	// Get session token first
	sessionResp, err := client.CallTool(ctx, "start_session", map[string]interface{}{})
	require.NoError(t, err, "Failed to call start_session")
	require.Nil(t, sessionResp.Error, "start_session returned error: %v", sessionResp.Error)

	var session TestSessionResponse
	parseToolResponse(t, sessionResp, &session)

	t.Run("ReadNonexistentFile", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "read_files", map[string]interface{}{
			"session_token": session.SessionToken,
			"paths":         []string{"/nonexistent/file.txt"},
		})
		require.NoError(t, err, "Failed to call read_files")
		require.Nil(t, resp.Error, "read_files should handle nonexistent files gracefully")

		// read_files should return with errors in the response, not as an MCP error
		var readResp TestReadFilesResponse
		parseToolResponse(t, resp, &readResp)
		assert.Greater(t, len(readResp.Errors), 0, "Should have errors for nonexistent file")
	})

	t.Run("CreateFileInNonAllowedPath", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "create_file", map[string]interface{}{
			"session_token": session.SessionToken,
			"path":          "/etc/should_not_work.txt",
			"content":       "This should fail",
		})
		require.NoError(t, err, "Failed to call create_file")
		assert.NotNil(t, resp.Error, "create_file should return error for non-allowed path")
	})
}

// parseToolResponse is a helper that parses MCP tool response content
func parseToolResponse(t *testing.T, resp *MCPResponse, target interface{}) {
	t.Helper()

	var result map[string]interface{}
	err := json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to parse tool response")

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Result should have content array")
	require.Greater(t, len(content), 0, "Content should not be empty")

	contentItem, ok := content[0].(map[string]interface{})
	require.True(t, ok, "Content item should be an object")

	text, ok := contentItem["text"].(string)
	require.True(t, ok, "Content item should have text")

	err = json.Unmarshal([]byte(text), target)
	require.NoError(t, err, "Failed to parse response content JSON")
}
