package test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test content templates for replace_file_part tests
const (
	GoTestContent = `package main

import "fmt"

const (
	ServerPort = "8080"
	AppName    = "test-app"
)

var (
	GlobalVar = "initial value"
	Counter   = 0
)

type Config struct {
	Port string
	Name string
}

type UserService interface {
	GetUser(id string) (*User, error)
}

func main() {
	fmt.Println("Hello, World!")
}

func oldFunction() string {
	return "old implementation"
}

func (c *Config) GetPort() string {
	return c.Port
}

func (*Config) SetDefaults() {
	// set defaults
}
`

	UpdatedFunction = `func newFunction() string {
	return "new implementation"
}`

	UpdatedType = `type Config struct {
	Port     string
	Name     string
	Version  string
	Features []string
}`

	UpdatedConst = `const (
	ServerPort = "9000"
	AppName    = "updated-app"
	Version    = "1.0.0"
)`

	UpdatedVar = `var (
	GlobalVar = "updated value"
	Counter   = 42
	NewVar    = "added variable"
)`

	UpdatedMethod = `func (c *Config) GetPort() string {
	if c.Port == "" {
		return "8080"
	}
	return c.Port
}`

	UpdatedInterface = `type UserService interface {
	GetUser(id string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id string) error
}`
)

// TestReplaceFilePart tests the replace_file_part tool functionality
func TestReplaceFilePart(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()
	testDir := GetTestDir()

	t.Run("ReplaceFunction", func(t *testing.T) {
		testReplaceFunction(t, client, ctx, testDir)
	})

	t.Run("ReplaceType", func(t *testing.T) {
		testReplaceType(t, client, ctx, testDir)
	})

	t.Run("ReplaceConst", func(t *testing.T) {
		testReplaceConst(t, client, ctx, testDir)
	})

	t.Run("ReplaceVar", func(t *testing.T) {
		testReplaceVar(t, client, ctx, testDir)
	})

	t.Run("ReplaceMethod", func(t *testing.T) {
		testReplaceMethod(t, client, ctx, testDir)
	})

	t.Run("ReplaceInterface", func(t *testing.T) {
		testReplaceInterface(t, client, ctx, testDir)
	})
}

func testReplaceFunction(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create test file
	testFilePath := filepath.Join(testDir, "replace_func_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the function
	resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "func",
		"part_name":   "oldFunction",
		"new_content": UpdatedFunction,
	})
	require.NoError(t, err, "Failed to call replace_file_part")
	require.Nil(t, resp.Error, "replace_file_part returned error: %v", resp.Error)

	// Verify the replacement
	updatedContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")

	content := string(updatedContent)
	assert.Contains(t, content, "func newFunction() string", "Should contain new function")
	assert.Contains(t, content, "new implementation", "Should contain new implementation")
	assert.NotContains(t, content, "func oldFunction()", "Should not contain old function")
	assert.NotContains(t, content, "old implementation", "Should not contain old implementation")

	// Verify other parts remain unchanged
	assert.Contains(t, content, "func main()", "Should preserve main function")
	assert.Contains(t, content, "func (c *Config) GetPort()", "Should preserve methods")
}

func testReplaceType(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create test file
	testFilePath := filepath.Join(testDir, "replace_type_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the type
	resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "type",
		"part_name":   "Config",
		"new_content": UpdatedType,
	})
	require.NoError(t, err, "Failed to call replace_file_part")
	require.Nil(t, resp.Error, "replace_file_part returned error: %v", resp.Error)

	// Verify the replacement
	updatedContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")

	content := string(updatedContent)
	assert.Contains(t, content, "Version  string", "Should contain new Version field")
	assert.Contains(t, content, "Features []string", "Should contain new Features field")

	// Verify interface type is preserved
	assert.Contains(t, content, "type UserService interface", "Should preserve UserService interface")
}

func testReplaceConst(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create test file
	testFilePath := filepath.Join(testDir, "replace_const_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the const block by targeting one of the constants
	resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "const",
		"part_name":   "ServerPort",
		"new_content": UpdatedConst,
	})
	require.NoError(t, err, "Failed to call replace_file_part")
	require.Nil(t, resp.Error, "replace_file_part returned error: %v", resp.Error)

	// Verify the replacement
	updatedContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")

	content := string(updatedContent)
	assert.Contains(t, content, `ServerPort = "9000"`, "Should contain updated ServerPort")
	assert.Contains(t, content, `AppName    = "updated-app"`, "Should contain updated AppName")
	assert.Contains(t, content, `Version    = "1.0.0"`, "Should contain new Version constant")
	assert.NotContains(t, content, `"8080"`, "Should not contain old port value")
}

func testReplaceVar(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create test file
	testFilePath := filepath.Join(testDir, "replace_var_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the var block by targeting one of the variables
	resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "var",
		"part_name":   "GlobalVar",
		"new_content": UpdatedVar,
	})
	require.NoError(t, err, "Failed to call replace_file_part")
	require.Nil(t, resp.Error, "replace_file_part returned error: %v", resp.Error)

	// Verify the replacement
	updatedContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")

	content := string(updatedContent)
	assert.Contains(t, content, `GlobalVar = "updated value"`, "Should contain updated GlobalVar")
	assert.Contains(t, content, `Counter   = 42`, "Should contain updated Counter")
	assert.Contains(t, content, `NewVar    = "added variable"`, "Should contain new variable")
	assert.NotContains(t, content, `"initial value"`, "Should not contain old value")
}

func testReplaceMethod(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create test file
	testFilePath := filepath.Join(testDir, "replace_method_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the method using receiver type syntax
	resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "func",
		"part_name":   "*Config.GetPort",
		"new_content": UpdatedMethod,
	})
	require.NoError(t, err, "Failed to call replace_file_part")
	require.Nil(t, resp.Error, "replace_file_part returned error: %v", resp.Error)

	// Verify the replacement
	updatedContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")

	content := string(updatedContent)
	assert.Contains(t, content, `if c.Port == ""`, "Should contain new logic")
	assert.Contains(t, content, `return "8080"`, "Should contain default return")

	// Verify other methods are preserved
	assert.Contains(t, content, "func (*Config) SetDefaults()", "Should preserve SetDefaults method")
	assert.Contains(t, content, "func main()", "Should preserve main function")
}

func testReplaceInterface(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create test file
	testFilePath := filepath.Join(testDir, "replace_interface_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the interface
	resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "type",
		"part_name":   "UserService",
		"new_content": UpdatedInterface,
	})
	require.NoError(t, err, "Failed to call replace_file_part")
	require.Nil(t, resp.Error, "replace_file_part returned error: %v", resp.Error)

	// Verify the replacement
	updatedContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")

	content := string(updatedContent)
	assert.Contains(t, content, "UpdateUser(user *User) error", "Should contain new UpdateUser method")
	assert.Contains(t, content, "DeleteUser(id string) error", "Should contain new DeleteUser method")
	assert.Contains(t, content, "GetUser(id string) (*User, error)", "Should preserve existing GetUser method")

	// Verify struct type is preserved
	assert.Contains(t, content, "type Config struct", "Should preserve Config struct")
}

// TestReplaceFilePartErrorCases tests error handling for replace_file_part
func TestReplaceFilePartErrorCases(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()
	testDir := GetTestDir()

	t.Run("UnsupportedLanguage", func(t *testing.T) {
		testFilePath := filepath.Join(testDir, "error_test.py")
		err := os.WriteFile(testFilePath, []byte("def hello():\n    print('hello')"), 0644)
		require.NoError(t, err, "Failed to create test file")

		resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
			"path":        testFilePath,
			"language":    "python",
			"part_type":   "func",
			"part_name":   "hello",
			"new_content": "def hello():\n    print('updated')",
		})
		require.NoError(t, err, "Failed to call replace_file_part")
		assert.NotNil(t, resp.Error, "Should return error for unsupported language")

		var content string
		parseToolResponse(t, resp, &content)
		assert.Contains(t, content, "language 'python' not supported", "Should mention unsupported language")
	})

	t.Run("InvalidPartType", func(t *testing.T) {
		testFilePath := filepath.Join(testDir, "error_test.go")
		err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
			"path":        testFilePath,
			"language":    "go",
			"part_type":   "invalid",
			"part_name":   "something",
			"new_content": "something",
		})
		require.NoError(t, err, "Failed to call replace_file_part")
		assert.NotNil(t, resp.Error, "Should return error for invalid part type")
	})

	t.Run("PartNotFound", func(t *testing.T) {
		testFilePath := filepath.Join(testDir, "not_found_test.go")
		err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
			"path":        testFilePath,
			"language":    "go",
			"part_type":   "func",
			"part_name":   "nonexistentFunction",
			"new_content": "func nonexistentFunction() {}",
		})
		require.NoError(t, err, "Failed to call replace_file_part")
		assert.NotNil(t, resp.Error, "Should return error for non-existent function")
	})

	t.Run("InvalidContentForType", func(t *testing.T) {
		testFilePath := filepath.Join(testDir, "invalid_content_test.go")
		err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
			"path":        testFilePath,
			"language":    "go",
			"part_type":   "func",
			"part_name":   "oldFunction",
			"new_content": "not a function", // Invalid content for func type
		})
		require.NoError(t, err, "Failed to call replace_file_part")
		assert.NotNil(t, resp.Error, "Should return error for invalid function content")
	})

	t.Run("InvalidGoSyntax", func(t *testing.T) {
		testFilePath := filepath.Join(testDir, "syntax_error_test.go")
		err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
			"path":        testFilePath,
			"language":    "go",
			"part_type":   "func",
			"part_name":   "oldFunction",
			"new_content": "func newFunction() { invalid syntax", // Invalid Go syntax
		})
		require.NoError(t, err, "Failed to call replace_file_part")
		assert.NotNil(t, resp.Error, "Should return error for invalid Go syntax")
	})

	t.Run("NonAllowedPath", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
			"path":        "/etc/hosts",
			"language":    "go",
			"part_type":   "func",
			"part_name":   "something",
			"new_content": "func something() {}",
		})
		require.NoError(t, err, "Failed to call replace_file_part")
		assert.NotNil(t, resp.Error, "Should return error for non-allowed path")
	})
}

// TestReplaceFilePartValidation tests input validation
func TestReplaceFilePartValidation(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()
	testDir := GetTestDir()

	// Create a valid test file
	testFilePath := filepath.Join(testDir, "validation_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	testCases := []struct {
		name        string
		partType    string
		newContent  string
		shouldError bool
		errorText   string
	}{
		{
			name:        "ValidFunc",
			partType:    "func",
			newContent:  "func validFunction() {}",
			shouldError: false,
		},
		{
			name:        "InvalidFunc",
			partType:    "func",
			newContent:  "not a function",
			shouldError: true,
			errorText:   "func replacement must start with 'func'",
		},
		{
			name:        "ValidType",
			partType:    "type",
			newContent:  "type ValidType struct { Field string }",
			shouldError: false,
		},
		{
			name:        "InvalidType",
			partType:    "type",
			newContent:  "not a type",
			shouldError: true,
			errorText:   "type replacement must start with 'type'",
		},
		{
			name:        "ValidConst",
			partType:    "const",
			newContent:  "MyConst = 42",
			shouldError: false,
		},
		{
			name:        "InvalidConst",
			partType:    "const",
			newContent:  "invalid const",
			shouldError: true,
			errorText:   "const replacement must contain '='",
		},
		{
			name:        "ValidVar",
			partType:    "var",
			newContent:  "MyVar = \"value\"",
			shouldError: false,
		},
		{
			name:        "InvalidVar",
			partType:    "var",
			newContent:  "invalid var",
			shouldError: true,
			errorText:   "var replacement must contain '='",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.CallTool(ctx, "replace_file_part", map[string]interface{}{
				"path":        testFilePath,
				"language":    "go",
				"part_type":   tc.partType,
				"part_name":   "ServerPort", // Use existing name for validation test
				"new_content": tc.newContent,
			})
			require.NoError(t, err, "Failed to call replace_file_part")

			if tc.shouldError {
				assert.NotNil(t, resp.Error, "Should return error for %s", tc.name)
				if tc.errorText != "" {
					var content string
					parseToolResponse(t, resp, &content)
					assert.Contains(t, content, tc.errorText, "Error should contain expected text")
				}
			} else {
				if resp.Error != nil {
					t.Logf("Unexpected error for %s: %v", tc.name, resp.Error)
				}
			}
		})
	}
}

// TestReplaceFilePartIntegration tests integration with the tools list
func TestReplaceFilePartIntegration(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()

	// Verify replace_file_part appears in tools list
	resp, err := client.ListTools(ctx)
	require.NoError(t, err, "Failed to list tools")
	require.Nil(t, resp.Error, "ListTools returned error: %v", resp.Error)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to parse tools list response")

	tools, ok := result["tools"].([]interface{})
	require.True(t, ok, "Tools should be an array")

	// Check that replace_file_part is in the list
	foundTool := false
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		require.True(t, ok, "Tool should be an object")

		name, ok := toolMap["name"].(string)
		require.True(t, ok, "Tool should have a name")

		if name == "replace_file_part" {
			foundTool = true

			// Verify tool description
			description, ok := toolMap["description"].(string)
			assert.True(t, ok, "replace_file_part should have a description")
			assert.Contains(t, description, "AST parsing", "Description should mention AST parsing")
			break
		}
	}

	assert.True(t, foundTool, "replace_file_part should be available in tools list")
}
