package test

import (
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

// TestReplaceFilePartDirect tests the replace_file_part tool functionality using direct server access
func TestReplaceFilePartDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("ReplaceFunction", func(t *testing.T) {
		testReplaceFunctionDirect(t, env)
	})

	t.Run("ReplaceType", func(t *testing.T) {
		testReplaceTypeDirect(t, env)
	})

	t.Run("ReplaceConst", func(t *testing.T) {
		testReplaceConstDirect(t, env)
	})

	t.Run("ReplaceVar", func(t *testing.T) {
		testReplaceVarDirect(t, env)
	})

	t.Run("ReplaceMethod", func(t *testing.T) {
		testReplaceMethodDirect(t, env)
	})

	t.Run("ReplaceInterface", func(t *testing.T) {
		testReplaceInterfaceDirect(t, env)
	})
}

func testReplaceFunctionDirect(t *testing.T, env *DirectServerTestEnv) {
	// Create test file
	testFilePath := filepath.Join(env.GetTestDir(), "replace_func_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the function
	result := env.CallTool(t, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "func",
		"part_name":   "oldFunction",
		"new_content": UpdatedFunction,
	})
	assert.NotNil(t, result, "replace_file_part should return result")

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

func testReplaceTypeDirect(t *testing.T, env *DirectServerTestEnv) {
	// Create test file
	testFilePath := filepath.Join(env.GetTestDir(), "replace_type_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the type
	result := env.CallTool(t, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "type",
		"part_name":   "Config",
		"new_content": UpdatedType,
	})
	assert.NotNil(t, result, "replace_file_part should return result")

	// Verify the replacement
	updatedContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")

	content := string(updatedContent)
	assert.Contains(t, content, "Version  string", "Should contain new Version field")
	assert.Contains(t, content, "Features []string", "Should contain new Features field")

	// Verify interface type is preserved
	assert.Contains(t, content, "type UserService interface", "Should preserve UserService interface")
}

func testReplaceConstDirect(t *testing.T, env *DirectServerTestEnv) {
	// Create test file
	testFilePath := filepath.Join(env.GetTestDir(), "replace_const_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the const block by targeting one of the constants
	result := env.CallTool(t, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "const",
		"part_name":   "ServerPort",
		"new_content": UpdatedConst,
	})
	assert.NotNil(t, result, "replace_file_part should return result")

	// Verify the replacement
	updatedContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")

	content := string(updatedContent)
	assert.Contains(t, content, `ServerPort = "9000"`, "Should contain updated ServerPort")
	assert.Contains(t, content, `AppName    = "updated-app"`, "Should contain updated AppName")
	assert.Contains(t, content, `Version    = "1.0.0"`, "Should contain new Version constant")
	assert.NotContains(t, content, `"8080"`, "Should not contain old port value")
}

func testReplaceVarDirect(t *testing.T, env *DirectServerTestEnv) {
	// Create test file
	testFilePath := filepath.Join(env.GetTestDir(), "replace_var_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the var block by targeting one of the variables
	result := env.CallTool(t, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "var",
		"part_name":   "GlobalVar",
		"new_content": UpdatedVar,
	})
	assert.NotNil(t, result, "replace_file_part should return result")

	// Verify the replacement
	updatedContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")

	content := string(updatedContent)
	assert.Contains(t, content, `GlobalVar = "updated value"`, "Should contain updated GlobalVar")
	assert.Contains(t, content, `Counter   = 42`, "Should contain updated Counter")
	assert.Contains(t, content, `NewVar    = "added variable"`, "Should contain new variable")
	assert.NotContains(t, content, `"initial value"`, "Should not contain old value")
}

func testReplaceMethodDirect(t *testing.T, env *DirectServerTestEnv) {
	// Create test file
	testFilePath := filepath.Join(env.GetTestDir(), "replace_method_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the method using receiver type syntax
	result := env.CallTool(t, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "func",
		"part_name":   "*Config.GetPort",
		"new_content": UpdatedMethod,
	})
	assert.NotNil(t, result, "replace_file_part should return result")

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

func testReplaceInterfaceDirect(t *testing.T, env *DirectServerTestEnv) {
	// Create test file
	testFilePath := filepath.Join(env.GetTestDir(), "replace_interface_test.go")
	err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Replace the interface
	result := env.CallTool(t, "replace_file_part", map[string]interface{}{
		"path":        testFilePath,
		"language":    "go",
		"part_type":   "type",
		"part_name":   "UserService",
		"new_content": UpdatedInterface,
	})
	assert.NotNil(t, result, "replace_file_part should return result")

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

// TestReplaceFilePartErrorCasesDirect tests error handling for replace_file_part
func TestReplaceFilePartErrorCasesDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("UnsupportedLanguage", func(t *testing.T) {
		testFilePath := filepath.Join(env.GetTestDir(), "error_test.py")
		err := os.WriteFile(testFilePath, []byte("def hello():\n    print('hello')"), 0644)
		require.NoError(t, err, "Failed to create test file")

		err = env.CallToolExpectError(t, "replace_file_part", map[string]interface{}{
			"path":        testFilePath,
			"language":    "python",
			"part_type":   "func",
			"part_name":   "hello",
			"new_content": "def hello():\n    print('updated')",
		})
		assert.Error(t, err, "Should return error for unsupported language")
		assert.Contains(t, err.Error(), "language 'python' not supported", "Should mention unsupported language")
	})

	t.Run("InvalidPartType", func(t *testing.T) {
		testFilePath := filepath.Join(env.GetTestDir(), "error_test.go")
		err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		err = env.CallToolExpectError(t, "replace_file_part", map[string]interface{}{
			"path":        testFilePath,
			"language":    "go",
			"part_type":   "invalid",
			"part_name":   "something",
			"new_content": "something",
		})
		assert.Error(t, err, "Should return error for invalid part type")
	})

	t.Run("PartNotFound", func(t *testing.T) {
		testFilePath := filepath.Join(env.GetTestDir(), "not_found_test.go")
		err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		err = env.CallToolExpectError(t, "replace_file_part", map[string]interface{}{
			"path":        testFilePath,
			"language":    "go",
			"part_type":   "func",
			"part_name":   "nonexistentFunction",
			"new_content": "func nonexistentFunction() {}",
		})
		assert.Error(t, err, "Should return error for non-existent function")
	})

	t.Run("InvalidContentForType", func(t *testing.T) {
		testFilePath := filepath.Join(env.GetTestDir(), "invalid_content_test.go")
		err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		err = env.CallToolExpectError(t, "replace_file_part", map[string]interface{}{
			"path":        testFilePath,
			"language":    "go",
			"part_type":   "func",
			"part_name":   "oldFunction",
			"new_content": "not a function", // Invalid content for func type
		})
		assert.Error(t, err, "Should return error for invalid function content")
	})

	t.Run("InvalidGoSyntax", func(t *testing.T) {
		testFilePath := filepath.Join(env.GetTestDir(), "syntax_error_test.go")
		err := os.WriteFile(testFilePath, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		err = env.CallToolExpectError(t, "replace_file_part", map[string]interface{}{
			"path":        testFilePath,
			"language":    "go",
			"part_type":   "func",
			"part_name":   "oldFunction",
			"new_content": "func newFunction() { invalid syntax", // Invalid Go syntax
		})
		assert.Error(t, err, "Should return error for invalid Go syntax")
	})

	t.Run("NonAllowedPath", func(t *testing.T) {
		err := env.CallToolExpectError(t, "replace_file_part", map[string]interface{}{
			"path":        "/etc/hosts",
			"language":    "go",
			"part_type":   "func",
			"part_name":   "something",
			"new_content": "func something() {}",
		})
		assert.Error(t, err, "Should return error for non-allowed path")
	})
}

// TestReplaceFilePartValidationDirect tests input validation using direct server access
func TestReplaceFilePartValidationDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	// Create a valid test file
	testFilePath := filepath.Join(env.GetTestDir(), "validation_test.go")
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
			newContent:  "func oldFunction() { return \"updated\" }",
			shouldError: false,
		},
		{
			name:        "InvalidFunc",
			partType:    "func",
			newContent:  "not a function",
			shouldError: true,
			errorText:   "func replacement must start with 'func '",
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
			errorText:   "type replacement must start with 'type '",
		},
		{
			name:        "ValidConst",
			partType:    "const",
			newContent:  "const (\n\tUpdatedServerPort = \"9000\"\n\tAppName = \"updated-app\"\n)",
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
			newContent:  "var (\n\tUpdatedGlobalVar = \"updated\"\n\tCounter = 42\n)",
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
			if tc.shouldError {
				// Choose appropriate part_name based on part_type
				partName := "ServerPort" // default for const
				if tc.partType == "func" {
					partName = "oldFunction"
				} else if tc.partType == "type" {
					partName = "Config"
				} else if tc.partType == "var" {
					partName = "GlobalVar"
				}

				err := env.CallToolExpectError(t, "replace_file_part", map[string]interface{}{
					"path":        testFilePath,
					"language":    "go",
					"part_type":   tc.partType,
					"part_name":   partName,
					"new_content": tc.newContent,
				})
				assert.Error(t, err, "Should return error for %s", tc.name)
				if tc.errorText != "" {
					assert.Contains(t, err.Error(), tc.errorText, "Error should contain expected text")
				}
			} else {
				// Choose appropriate part_name based on part_type
				partName := "ServerPort" // default for const
				if tc.partType == "func" {
					partName = "oldFunction"
				} else if tc.partType == "type" {
					partName = "Config"
				} else if tc.partType == "var" {
					partName = "GlobalVar"
				}

				result := env.CallTool(t, "replace_file_part", map[string]interface{}{
					"path":        testFilePath,
					"language":    "go",
					"part_type":   tc.partType,
					"part_name":   partName,
					"new_content": tc.newContent,
				})
				assert.NotNil(t, result, "Should succeed for valid %s", tc.name)
			}
		})
	}
}
