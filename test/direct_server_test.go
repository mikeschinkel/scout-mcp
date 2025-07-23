package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DirectServerTestEnv provides direct access to MCP server and tools for testing
type DirectServerTestEnv struct {
	server       *scout.MCPServer
	testDir      string
	ctx          context.Context
	cleanup      func()
	sessionToken string
}

// NewDirectServerTestEnv creates a test environment with direct MCP server access
func NewDirectServerTestEnv(t *testing.T) *DirectServerTestEnv {
	t.Helper()

	// Initialize loggers with quiet logger for tests
	quietLogger := testutil.QuietLogger()
	scout.SetLogger(quietLogger)
	mcptools.SetLogger(quietLogger)

	// Create temporary test directory
	testDir, err := os.MkdirTemp("", "scout-mcp-direct-test-")
	require.NoError(t, err, "Failed to create test directory")

	// Create some test files
	testFiles := map[string]string{
		"README.md":       "# Test Project\nThis is a test project for Scout MCP testing.\n",
		"src/main.go":     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
		"src/util.go":     "package main\n\nfunc helper() string {\n\treturn \"helper\"\n}\n",
		"docs/guide.md":   "# User Guide\nThis is the user guide.\n",
		"config.json":     "{\"version\": \"1.0\", \"debug\": true}\n",
		"data/sample.txt": "Sample data file\nWith multiple lines\n",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(testDir, filePath)

		// Create parent directories
		err = os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err, "Failed to create parent directory")

		// Create file
		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")
	}

	// Create MCP server directly
	opts := scout.Opts{
		OnlyMode:        true, // Use only the test directory
		AdditionalPaths: []string{testDir},
	}

	server, err := scout.NewMCPServer(opts)
	require.NoError(t, err, "Failed to create MCP server")

	ctx := context.Background()

	// Get session token by calling start_session tool directly
	startSessionTool := mcputil.GetRegisteredTool("start_session")
	require.NotNil(t, startSessionTool, "start_session tool should be available")

	// Create empty request for start_session (it doesn't need parameters)
	req := testutil.NewMockRequest(map[string]interface{}{})

	result, err := startSessionTool.Handle(ctx, req)
	require.NoError(t, err, "start_session should succeed")

	// Extract session token from result
	sessionToken, err := extractSessionToken(t, result)
	require.NoError(t, err, "Should be able to extract session token")
	require.NotEmpty(t, sessionToken, "Session token should not be empty")

	// Setup cleanup
	cleanup := func() {
		if testDir != "" {
			testutil.MaybeRemove(t, testDir)
		}
	}

	return &DirectServerTestEnv{
		server:       server,
		testDir:      testDir,
		ctx:          ctx,
		cleanup:      cleanup,
		sessionToken: sessionToken,
	}
}

// CallTool calls a tool directly with the given arguments
func (env *DirectServerTestEnv) CallTool(t *testing.T, toolName string, args map[string]interface{}) mcputil.ToolResult {
	t.Helper()

	// Add session token if not present
	if _, exists := args["session_token"]; !exists && toolName != "start_session" {
		args["session_token"] = env.sessionToken
	}

	// Get the tool
	tool := mcputil.GetRegisteredTool(toolName)
	require.NotNil(t, tool, "Tool %s should be registered", toolName)

	// Create request
	req := testutil.NewMockRequest(args)

	// Check preconditions first (like the real MCP server does)
	err := tool.EnsurePreconditions(req)
	if err != nil {
		require.NoError(t, err, "Tool %s preconditions failed", toolName)
	}

	// Call the tool
	result, err := tool.Handle(env.ctx, req)
	require.NoError(t, err, "Tool %s should not error", toolName)

	return result
}

// CallToolExpectError calls a tool and expects it to return an error
func (env *DirectServerTestEnv) CallToolExpectError(t *testing.T, toolName string, args map[string]interface{}) error {
	t.Helper()

	// Add session token if not present
	if _, exists := args["session_token"]; !exists && toolName != "start_session" {
		args["session_token"] = env.sessionToken
	}

	// Get the tool
	tool := mcputil.GetRegisteredTool(toolName)
	require.NotNil(t, tool, "Tool %s should be registered", toolName)

	// Create request
	req := testutil.NewMockRequest(args)

	// Check preconditions first (like the real MCP server does)
	err := tool.EnsurePreconditions(req)
	if err != nil {
		// Precondition failed - this is the expected error
		return err
	}

	// Call the tool
	result, err := tool.Handle(env.ctx, req)
	if err != nil {
		// Go error returned
		return err
	}

	// Check if result is an errorResult using reflection
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() == reflect.Ptr {
		resultValue = resultValue.Elem()
	}

	// Check if it's an errorResult
	if resultValue.Type().Name() == "errorResult" {
		// Extract error message
		messageField := resultValue.FieldByName("message")
		if messageField.IsValid() {
			errorMsg := messageField.String()
			return fmt.Errorf("%s", errorMsg)
		}
	}

	// No error found - this is unexpected
	require.Fail(t, "Tool %s should return error but returned success", toolName)
	return nil
}

// GetTestDir returns the test directory path
func (env *DirectServerTestEnv) GetTestDir() string {
	return env.testDir
}

// GetSessionToken returns the session token
func (env *DirectServerTestEnv) GetSessionToken() string {
	return env.sessionToken
}

// Cleanup cleans up test resources
func (env *DirectServerTestEnv) Cleanup() {
	if env.cleanup != nil {
		env.cleanup()
	}
}

// extractSessionToken extracts the session token from start_session result
func extractSessionToken(t *testing.T, result mcputil.ToolResult) (string, error) {
	t.Helper()

	// Use reflection to extract the text from textResult
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() == reflect.Ptr {
		resultValue = resultValue.Elem()
	}

	textField := resultValue.FieldByName("text")
	if !textField.IsValid() {
		return "", assert.AnError
	}

	jsonText := textField.String()

	// Parse the JSON content
	var response struct {
		SessionToken string `json:"session_token"`
	}

	err := json.Unmarshal([]byte(jsonText), &response)
	if err != nil {
		return "", err
	}

	return response.SessionToken, nil
}

// ParseJSONResult extracts and parses JSON from a ToolResult
func ParseJSONResult(t *testing.T, result mcputil.ToolResult, target interface{}) {
	t.Helper()

	// Use reflection to extract the text from textResult
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() == reflect.Ptr {
		resultValue = resultValue.Elem()
	}

	textField := resultValue.FieldByName("text")
	require.True(t, textField.IsValid(), "Result should have text field")

	jsonText := textField.String()

	// Parse the JSON content
	err := json.Unmarshal([]byte(jsonText), target)
	require.NoError(t, err, "Should be able to parse JSON result")
}

// ParseTextResult extracts text from a ToolResult
func ParseTextResult(t *testing.T, result mcputil.ToolResult) string {
	t.Helper()

	// Use reflection to extract the text from textResult
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() == reflect.Ptr {
		resultValue = resultValue.Elem()
	}

	textField := resultValue.FieldByName("text")
	require.True(t, textField.IsValid(), "Result should have text field")

	return textField.String()
}

// Test the direct server infrastructure
func TestDirectServerInfrastructure(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("ServerCreated", func(t *testing.T) {
		assert.NotNil(t, env.server, "Server should be created")
		assert.NotEmpty(t, env.GetTestDir(), "Test directory should be set")
		assert.NotEmpty(t, env.GetSessionToken(), "Session token should be available")
	})

	t.Run("ToolsRegistered", func(t *testing.T) {
		expectedTools := []string{
			"start_session", "read_files", "search_files", "get_config",
			"tool_help", "create_file", "update_file", "delete_files",
		}

		for _, toolName := range expectedTools {
			tool := mcputil.GetRegisteredTool(toolName)
			assert.NotNil(t, tool, "Tool %s should be registered", toolName)
		}
	})

	t.Run("BasicToolCall", func(t *testing.T) {
		// Test get_config tool
		result := env.CallTool(t, "get_config", map[string]interface{}{})
		assert.NotNil(t, result, "get_config should return result")
	})

	t.Run("FileOperations", func(t *testing.T) {
		// Test creating a file
		testFile := filepath.Join(env.GetTestDir(), "new_test_file.txt")

		result := env.CallTool(t, "create_file", map[string]interface{}{
			"filepath":    testFile,
			"new_content": "Test file content",
			"create_dirs": true,
		})
		assert.NotNil(t, result, "create_file should succeed")

		// Verify file was created
		content, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read created file")
		assert.Equal(t, "Test file content", string(content), "File content should match")
	})

	t.Run("ReadFiles", func(t *testing.T) {
		// Test reading existing files
		result := env.CallTool(t, "read_files", map[string]interface{}{
			"paths": []string{
				filepath.Join(env.GetTestDir(), "README.md"),
				filepath.Join(env.GetTestDir(), "src/main.go"),
			},
		})
		assert.NotNil(t, result, "read_files should succeed")
	})
}
