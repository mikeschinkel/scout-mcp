package mcptools_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("get_config")
	require.NotNil(t, tool, "get_config tool should be registered")

	tool.SetConfig(config)

	t.Run("GetBasicConfig", func(t *testing.T) {
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error getting config")
		assert.NotNil(t, result, "Result should not be nil")

		// The result should be a JSON structure with config information
		// We can't easily parse it here without knowing the exact format,
		// but we can verify it's not nil and no error occurred
	})

}

func TestToolHelpTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("tool_help")
	require.NotNil(t, tool, "tool_help tool should be registered")

	tool.SetConfig(config)

	t.Run("GetFullDocumentation", func(t *testing.T) {
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error getting full documentation")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("GetSpecificToolHelp", func(t *testing.T) {
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"tool":          "read_files",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error getting specific tool help")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("GetHelpForNonExistentTool", func(t *testing.T) {
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"tool":          "nonexistent_tool",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error when tool not found")
		assert.NotNil(t, result, "Result should not be nil")
		// The tool should return helpful message about available tools
	})
}

func TestAnalyzeFilesTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("analyze_files")
	require.NotNil(t, tool, "analyze_files tool should be registered")

	tool.SetConfig(config)

	t.Run("AnalyzeSingleFile", func(t *testing.T) {
		// Create a more complex file for analysis
		testFile := filepath.Join(tempDir, "analyze_test.go")
		complexContent := `package main

import (
	"fmt"
	"os"
	"net/http"
)

const (
	DefaultPort = "8080"
	MaxRetries  = 3
)

var (
	config Config
	logger Logger
)

type Config struct {
	Port     string
	Host     string
	Database DatabaseConfig
}

type DatabaseConfig struct {
	URL      string
	Username string
	Password string
}

func main() {
	server := &http.Server{
		Addr:    ":" + DefaultPort,
		Handler: setupRoutes(),
	}
	
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server failed: %v\n", err)
		os.Exit(1)
	}
}

func setupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/health", healthHandler)
	return mux
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the server!")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}
`
		err := os.WriteFile(testFile, []byte(complexContent), 0644)
		require.NoError(t, err, "Failed to create complex test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"files":         []any{testFile},
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error analyzing file")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("AnalyzeMultipleFiles", func(t *testing.T) {
		// Create multiple files
		file1 := filepath.Join(tempDir, "simple.txt")
		err := os.WriteFile(file1, []byte("Simple text file\nWith two lines"), 0644)
		require.NoError(t, err, "Failed to create simple file")

		file2 := filepath.Join(tempDir, "config.json")
		configContent := `{
  "server": {
    "port": 8080,
    "host": "localhost"
  },
  "database": {
    "url": "postgres://localhost/mydb",
    "maxConnections": 10
  }
}`
		err = os.WriteFile(file2, []byte(configContent), 0644)
		require.NoError(t, err, "Failed to create config file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"files":         []any{file1, file2},
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error analyzing multiple files")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("AnalyzeNonExistentFile", func(t *testing.T) {
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"files":         []any{"/nonexistent/file.txt"},
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error with non-existent file")
		assert.NotNil(t, result, "Result should not be nil")
		// The tool should handle missing files gracefully and report them
	})
}

func TestRequestApprovalTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("request_approval")
	require.NotNil(t, tool, "request_approval tool should be registered")

	tool.SetConfig(config)

	t.Run("RequestBasicApproval", func(t *testing.T) {
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token":   token,
			"operation":       "create_file",
			"files":           []any{filepath.Join(tempDir, "new_file.txt")},
			"impact_summary":  "Creating a new configuration file",
			"risk_level":      "low",
			"preview_content": "# Configuration\nport: 8080\nhost: localhost",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error requesting approval")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("RequestHighRiskApproval", func(t *testing.T) {
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token":   token,
			"operation":       "delete_files",
			"files":           []any{filepath.Join(tempDir, "important_file.txt")},
			"impact_summary":  "Deleting critical system file",
			"risk_level":      "high",
			"preview_content": "This operation will permanently delete the file",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error requesting high-risk approval")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("InvalidRiskLevel", func(t *testing.T) {
		// Note: request_approval tool is currently a stub implementation
		// that doesn't validate risk levels yet
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token":  token,
			"operation":      "update_file",
			"files":          []any{filepath.Join(tempDir, "test.txt")},
			"impact_summary": "Updating file",
			"risk_level":     "invalid_level",
		})

		result, err := testutil.CallTool(tool, req)
		// TODO: When tool is fully implemented, this should return an error
		require.NoError(t, err, "Stub implementation doesn't validate risk levels yet")
		assert.NotNil(t, result, "Result should not be nil")
	})
}

func TestGenerateApprovalTokenTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("generate_approval_token")
	require.NotNil(t, tool, "generate_approval_token tool should be registered")

	tool.SetConfig(config)

	t.Run("GenerateTokenForFileOperations", func(t *testing.T) {
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"file_actions": []any{
				mcptools.FileAction{Action: "create", Path: "/test/file1.txt", Purpose: "test creation"},
				mcptools.FileAction{Action: "update", Path: "/test/file2.txt", Purpose: "test update"},
			},
			"operations": []any{"create_file", "update_file"},
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error generating approval token")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("GenerateTokenForDeleteOperations", func(t *testing.T) {
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"file_actions": []any{
				mcptools.FileAction{Action: "delete", Path: "/test/file.txt", Purpose: "test deletion"},
			},
			"operations": []any{"delete_files"},
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error generating delete approval token")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("MissingRequiredParameters", func(t *testing.T) {
		// Note: The tool treats missing file_actions and operations as empty arrays,
		// which is valid behavior - it generates a token for empty action lists
		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			// Missing file_actions and operations - treated as empty arrays
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Tool accepts empty file_actions and operations")
		assert.NotNil(t, result, "Result should not be nil")
	})
}

// Test tool metadata and registration
func TestToolMetadata(t *testing.T) {
	expectedTools := map[string]struct{}{
		"start_session":           {},
		"read_files":              {},
		"search_files":            {},
		"get_config":              {},
		"tool_help":               {},
		"create_file":             {},
		"update_file":             {},
		"delete_files":            {},
		"update_file_lines":       {},
		"delete_file_lines":       {},
		"insert_file_lines":       {},
		"insert_at_pattern":       {},
		"replace_pattern":         {},
		"find_file_part":          {},
		"replace_file_part":       {},
		"validate_files":          {},
		"analyze_files":           {},
		"request_approval":        {},
		"generate_approval_token": {},
	}

	// Get all registered tools
	registeredTools := mcputil.RegisteredToolsMap()

	t.Run("AllExpectedToolsRegistered", func(t *testing.T) {
		for expectedTool := range expectedTools {
			_, exists := registeredTools[expectedTool]
			assert.True(t, exists, "Expected tool %s should be registered", expectedTool)
		}
	})

	t.Run("NoUnexpectedTools", func(t *testing.T) {
		for registeredTool := range registeredTools {
			_, expected := expectedTools[registeredTool]
			assert.True(t, expected, "Unexpected tool %s is registered", registeredTool)
		}
	})

	t.Run("ToolCount", func(t *testing.T) {
		assert.Equal(t, len(expectedTools), len(registeredTools),
			"Number of registered tools should match expected count")
	})

	t.Run("ToolBasicProperties", func(t *testing.T) {
		for toolName := range expectedTools {
			tool := mcputil.GetRegisteredTool(toolName)
			require.NotNil(t, tool, "Tool %s should be accessible", toolName)

			assert.Equal(t, toolName, tool.Name(), "Tool name should match")
			assert.NotEmpty(t, tool.Options().Description, "Tool %s should have description", toolName)

			// All tools except start_session should have session_token property
			if toolName != "start_session" {
				hasSessionToken := false
				for _, prop := range tool.Options().Properties {
					if prop.GetName() == "session_token" {
						hasSessionToken = true
						break
					}
				}
				assert.True(t, hasSessionToken, "Tool %s should have session_token property", toolName)
			}
		}
	})
}
