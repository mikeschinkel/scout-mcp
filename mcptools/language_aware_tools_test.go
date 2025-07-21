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

	UpdatedInterface = `type UserService interface {
	GetUser(id string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id string) error
}`

	UpdatedMethod = `func (c *Config) GetPort() string {
	if c.Port == "" {
		return "8080"
	}
	return c.Port
}`
)

func TestFindFilePartTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("find_file_part")
	require.NotNil(t, tool, "find_file_part tool should be registered")

	tool.SetConfig(config)

	t.Run("FindGoFunction", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "find_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "oldFunction",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error finding function")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("FindGoType", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "find_type_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "type",
			"part_name":     "Config",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error finding type")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("FindGoConst", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "find_const_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "const",
			"part_name":     "ServerPort",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error finding const")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("FindGoVar", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "find_var_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "var",
			"part_name":     "GlobalVar",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error finding var")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("FindMethod", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "find_method_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "*Config.GetPort",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error finding method")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("PartNotFound", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "find_missing_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "nonexistentFunction",
		})

		result, err := testutil.CallTool(tool, req)
		assert.Error(t, err, "Should error when part not found")
		assert.Nil(t, result, "Result should be nil on error")
	})
}

func TestReplaceFilePartTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("replace_file_part")
	require.NotNil(t, tool, "replace_file_part tool should be registered")

	tool.SetConfig(config)

	t.Run("ReplaceFunction", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "replace_func_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "oldFunction",
			"new_content":   UpdatedFunction,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error replacing function")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the replacement
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		content := string(updatedContent)
		assert.Contains(t, content, "func newFunction() string", "Should contain new function")
		assert.Contains(t, content, "new implementation", "Should contain new implementation")
		assert.NotContains(t, content, "old implementation", "Should not contain old implementation")
	})

	t.Run("ReplaceType", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "replace_type_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "type",
			"part_name":     "Config",
			"new_content":   UpdatedType,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error replacing type")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the replacement
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		content := string(updatedContent)
		assert.Contains(t, content, "Version  string", "Should contain new Version field")
		assert.Contains(t, content, "Features []string", "Should contain new Features field")
	})

	t.Run("ReplaceConst", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "replace_const_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "const",
			"part_name":     "ServerPort",
			"new_content":   UpdatedConst,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error replacing const")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the replacement
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		content := string(updatedContent)
		assert.Contains(t, content, `ServerPort = "9000"`, "Should contain updated ServerPort")
		assert.Contains(t, content, `Version    = "1.0.0"`, "Should contain new Version constant")
	})

	t.Run("ReplaceVar", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "replace_var_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "var",
			"part_name":     "GlobalVar",
			"new_content":   UpdatedVar,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error replacing var")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the replacement
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		content := string(updatedContent)
		assert.Contains(t, content, `GlobalVar = "updated value"`, "Should contain updated GlobalVar")
		assert.Contains(t, content, `NewVar    = "added variable"`, "Should contain new variable")
	})

	t.Run("ReplaceInterface", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "replace_interface_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "type",
			"part_name":     "UserService",
			"new_content":   UpdatedInterface,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error replacing interface")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the replacement
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		content := string(updatedContent)
		assert.Contains(t, content, "UpdateUser(user *User) error", "Should contain new UpdateUser method")
		assert.Contains(t, content, "DeleteUser(id string) error", "Should contain new DeleteUser method")
	})

	t.Run("ReplaceMethod", func(t *testing.T) {
		// Create Go test file
		testFile := filepath.Join(tempDir, "replace_method_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "*Config.GetPort",
			"new_content":   UpdatedMethod,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error replacing method")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the replacement
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		content := string(updatedContent)
		assert.Contains(t, content, `if c.Port == ""`, "Should contain new logic")
		assert.Contains(t, content, `return "8080"`, "Should contain default return")
	})

	t.Run("UnsupportedLanguage", func(t *testing.T) {
		// Create Python file (unsupported language)
		testFile := filepath.Join(tempDir, "unsupported.py")
		err := os.WriteFile(testFile, []byte("def hello():\n    print('hello')"), 0644)
		require.NoError(t, err, "Failed to create Python test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "python",
			"part_type":     "func",
			"part_name":     "hello",
			"new_content":   "def hello():\n    print('updated')",
		})

		result, err := testutil.CallTool(tool, req)
		assert.Error(t, err, "Should error for unsupported language")
		assert.Nil(t, result, "Result should be nil on error")
	})

	t.Run("InvalidPartType", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "invalid_part_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "invalid",
			"part_name":     "something",
			"new_content":   "something",
		})

		result, err := testutil.CallTool(tool, req)
		assert.Error(t, err, "Should error for invalid part type")
		assert.Nil(t, result, "Result should be nil on error")
	})

	t.Run("PartNotFound", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "not_found_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "nonexistentFunction",
			"new_content":   "func nonexistentFunction() {}",
		})

		result, err := testutil.CallTool(tool, req)
		assert.Error(t, err, "Should error for non-existent function")
		assert.Nil(t, result, "Result should be nil on error")
	})
}

func TestValidateFilesTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("validate_files")
	require.NotNil(t, tool, "validate_files tool should be registered")

	tool.SetConfig(config)

	t.Run("ValidateValidGoFile", func(t *testing.T) {
		// Create valid Go file
		testFile := filepath.Join(tempDir, "valid_test.go")
		err := os.WriteFile(testFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create valid Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"files":         []any{testFile},
			"paths":         []any{tempDir},
			"language":      "go",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error validating valid Go file")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("ValidateInvalidGoFile", func(t *testing.T) {
		// Create invalid Go file
		testFile := filepath.Join(tempDir, "invalid_test.go")
		invalidContent := "package main\n\nfunc main() {\n    invalid syntax here"
		err := os.WriteFile(testFile, []byte(invalidContent), 0644)
		require.NoError(t, err, "Failed to create invalid Go test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"files":         []any{testFile},
			"paths":         []any{tempDir},
			"language":      "go",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error when validating invalid file")
		assert.NotNil(t, result, "Result should not be nil")
		// The tool should return validation errors in the result, not as an error
	})

	t.Run("ValidateDirectory", func(t *testing.T) {
		// Create multiple Go files
		validFile := filepath.Join(tempDir, "valid.go")
		err := os.WriteFile(validFile, []byte(GoTestContent), 0644)
		require.NoError(t, err, "Failed to create valid Go file")

		invalidFile := filepath.Join(tempDir, "invalid.go")
		invalidContent := "package main\n\nfunc main() { invalid"
		err = os.WriteFile(invalidFile, []byte(invalidContent), 0644)
		require.NoError(t, err, "Failed to create invalid Go file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"files":         []any{},
			"paths":         []any{tempDir},
			"language":      "go",
			"extensions":    []any{".go"},
			"recursive":     true,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error validating directory")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("UnsupportedLanguage", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.py")
		err := os.WriteFile(testFile, []byte("print('hello')"), 0644)
		require.NoError(t, err, "Failed to create Python test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"files":         []any{testFile},
			"paths":         []any{tempDir},
			"language":      "python",
		})

		result, err := testutil.CallTool(tool, req)
		// Tool doesn't error for unsupported language, but returns result with error info
		require.NoError(t, err, "Tool should not error for unsupported language")
		assert.NotNil(t, result, "Result should not be nil")
		// The tool handles unsupported languages gracefully by returning status info
	})
}
