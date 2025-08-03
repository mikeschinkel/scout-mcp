package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
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
	testLogger := testutil.NewTestLogger()
	scout.SetLogger(testLogger)
	mcptools.SetLogger(testLogger)
	mcputil.SetLogger(testLogger)

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
	req := mcputil.NewMockRequest(map[string]interface{}{})

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

// extractSessionToken extracts the session token from start_session result
func extractSessionToken(t *testing.T, result mcputil.ToolResult) (string, error) {
	t.Helper()

	// Get the JSON text from the result
	jsonText := result.Value()

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

// Cleanup cleans up test resources
func (env *DirectServerTestEnv) Cleanup() {
	if env.cleanup != nil {
		env.cleanup()
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
	req := mcputil.NewMockRequest(args)

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
	req := mcputil.NewMockRequest(args)

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

	// Try to get the result value - if it's an error result, it will contain the error message
	resultValue := result.Value()
	if resultValue != "" {
		// This might be an error result or a successful result
		// For error results, the Value() contains the error message directly
		// For successful results, it contains JSON

		// Try to parse as JSON first - if it fails, assume it's an error message
		var temp interface{}
		err := json.Unmarshal([]byte(resultValue), &temp)
		if err != nil {
			// Not valid JSON, likely an error message
			return fmt.Errorf("%s", resultValue)
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
