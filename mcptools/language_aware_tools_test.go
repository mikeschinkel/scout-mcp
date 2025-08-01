package mcptools_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const LanguageAwareDirPrefix = "language-aware-test"

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

// Language-aware tool result types
type LanguageAwareResult struct {
	// Common fields
	Success  bool   `json:"success,omitempty"`
	FilePath string `json:"file_path,omitempty"`
	Message  string `json:"message,omitempty"`
	// Find operations specific
	Found       bool   `json:"found,omitempty"`
	StartLine   int    `json:"start_line,omitempty"`
	EndLine     int    `json:"end_line,omitempty"`
	StartOffset int    `json:"start_offset,omitempty"`
	EndOffset   int    `json:"end_offset,omitempty"`
	Content     string `json:"content,omitempty"`
	// Find/Replace operations
	Language string `json:"language,omitempty"`
	PartType string `json:"part_type,omitempty"`
	PartName string `json:"part_name,omitempty"`
	// Validation operations
	TotalFiles   int                `json:"total_files,omitempty"`
	ValidFiles   int                `json:"valid_files,omitempty"`
	InvalidFiles int                `json:"invalid_files,omitempty"`
	OverallValid bool               `json:"overall_valid,omitempty"`
	Results      []ValidationResult `json:"results,omitempty"`
}

type ValidationResult struct {
	FilePath string `json:"file_path"`
	Language string `json:"language"`
	Valid    bool   `json:"valid"`
	Error    string `json:"error,omitempty"`
}

type LanguageAwareResultOpts struct {
	ExpectError       bool
	ExpectedErrorMsg  string
	ExpectedFilePath  string
	ExpectedMessage   string
	ExpectedSuccess   bool
	ExpectedFound     bool
	ShouldUpdateFile  bool
	ShouldContainText string
	ShouldNotContain  string
	ExpectedContent   string
	// Find/Replace operations
	ExpectedLanguage    string
	ExpectedPartType    string
	ExpectedPartName    string
	ExpectedStartLine   int
	ExpectedEndLine     int
	ExpectedStartOffset int
	ExpectedEndOffset   int
	// Validation operations
	ExpectedTotalFiles    int
	ExpectedValidFiles    int
	ExpectedInvalidFiles  int
	ExpectedOverallValid  bool
	ExpectedValidation    bool
	CheckValidationErrors bool
}

func requireLanguageAwareResult(t *testing.T, result *LanguageAwareResult, err error, opts LanguageAwareResultOpts) {
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

	// Check find/replace tool result structure
	if opts.ExpectedSuccess {
		assert.True(t, result.Success, "Operation should be successful")
	}

	if opts.ExpectedFound {
		assert.True(t, result.Found, "Item should be found")
	}

	if opts.ExpectedFilePath != "" {
		assert.Equal(t, opts.ExpectedFilePath, result.FilePath, "File path should match expected")
	}

	if opts.ExpectedLanguage != "" {
		assert.Equal(t, opts.ExpectedLanguage, result.Language, "Language should match expected")
	}

	if opts.ExpectedPartType != "" {
		assert.Equal(t, opts.ExpectedPartType, result.PartType, "Part type should match expected")
	}

	if opts.ExpectedPartName != "" {
		assert.Equal(t, opts.ExpectedPartName, result.PartName, "Part name should match expected")
	}

	if opts.ExpectedStartLine > 0 {
		assert.Equal(t, opts.ExpectedStartLine, result.StartLine, "Start line should match expected")
	}

	if opts.ExpectedEndLine > 0 {
		assert.Equal(t, opts.ExpectedEndLine, result.EndLine, "End line should match expected")
	}

	// Check validation tool result structure
	if opts.ExpectedTotalFiles > 0 {
		assert.Equal(t, opts.ExpectedTotalFiles, result.TotalFiles, "Total files should match expected")
	}

	if opts.ExpectedValidFiles >= 0 {
		assert.Equal(t, opts.ExpectedValidFiles, result.ValidFiles, "Valid files should match expected")
	}

	if opts.ExpectedInvalidFiles >= 0 {
		assert.Equal(t, opts.ExpectedInvalidFiles, result.InvalidFiles, "Invalid files should match expected")
	}

	assert.Equal(t, opts.ExpectedOverallValid, result.OverallValid, "Overall valid should match expected")

	// Check validation results array
	if len(result.Results) > 0 {
		assert.Greater(t, len(result.Results), 0, "Should have validation results")

		if opts.CheckValidationErrors {
			// Check for validation errors in the results
			foundErrors := false
			for _, fileResult := range result.Results {
				if fileResult.Error != "" {
					foundErrors = true
					break
				}
			}
			assert.True(t, foundErrors, "Should have validation errors in results")
		}

		// Check individual file validation if specified
		if len(result.Results) > 0 {
			firstFile := result.Results[0]
			if opts.ExpectedValidation {
				assert.True(t, firstFile.Valid, "First file should be valid")
			} else if !opts.ExpectedValidation && opts.ExpectedTotalFiles > 0 {
				// If we expect specific validation state, check it
				assert.False(t, firstFile.Valid, "File should be invalid")
			}
		}
	}

	// Check file system side effects
	if opts.ShouldUpdateFile && opts.ExpectedFilePath != "" {
		_, err := os.Stat(opts.ExpectedFilePath)
		assert.NoError(t, err, "File should exist on disk: %s", opts.ExpectedFilePath)

		if opts.ExpectedContent != "" {
			content, readErr := os.ReadFile(opts.ExpectedFilePath)
			require.NoError(t, readErr, "Should be able to read updated file")
			assert.Equal(t, opts.ExpectedContent, string(content), "File content should match expected")
		}

		if opts.ShouldContainText != "" {
			content, readErr := os.ReadFile(opts.ExpectedFilePath)
			require.NoError(t, readErr, "Should be able to read updated file")
			assert.Contains(t, string(content), opts.ShouldContainText, "File should contain expected text")
		}

		if opts.ShouldNotContain != "" {
			content, readErr := os.ReadFile(opts.ExpectedFilePath)
			require.NoError(t, readErr, "Should be able to read updated file")
			assert.NotContains(t, string(content), opts.ShouldNotContain, "File should not contain specified text")
		}
	}
}

func TestFindFilePartTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("find_file_part")
	require.NotNil(t, tool, "find_file_part tool should be registered")

	t.Run("FindGoFunction_ShouldLocateAndReturnFunctionDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("find-func-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "oldFunction",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error finding function",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedFound:    true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedPartType: "func",
			ExpectedPartName: "oldFunction",
		})
	})

	t.Run("FindGoType_ShouldLocateAndReturnTypeDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("find-type-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_type_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "type",
			"part_name":     "Config",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error finding type",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedFound:    true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedPartType: "type",
			ExpectedPartName: "Config",
		})
	})

	t.Run("FindGoConst_ShouldLocateAndReturnConstDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("find-const-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_const_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "const",
			"part_name":     "ServerPort",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error finding const",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedFound:    true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedPartType: "const",
			ExpectedPartName: "ServerPort",
		})
	})

	t.Run("FindGoVar_ShouldLocateAndReturnVarDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("find-var-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_var_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "var",
			"part_name":     "GlobalVar",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error finding var",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedFound:    true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedPartType: "var",
			ExpectedPartName: "GlobalVar",
		})
	})

	t.Run("FindMethod_ShouldLocateAndReturnMethodDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("find-method-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_method_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "*Config.GetPort",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error finding method",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedFound:    true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedPartType: "func",
			ExpectedPartName: "*Config.GetPort",
		})
	})

	t.Run("PartNotFound_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("missing-part-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_missing_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "nonexistentFunction",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle part not found",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "not found",
		})
	})
}

func TestReplaceFilePartTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("replace_file_part")
	require.NotNil(t, tool, "replace_file_part tool should be registered")

	t.Run("ReplaceFunction_ShouldUpdateFunctionAndReturnDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-func-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_func_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "oldFunction",
			"new_content":   UpdatedFunction,
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error replacing function",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedSuccess:   true,
			ExpectedFilePath:  testFile.Filepath,
			ExpectedLanguage:  "go",
			ExpectedPartType:  "func",
			ExpectedPartName:  "oldFunction",
			ShouldUpdateFile:  true,
			ShouldContainText: "func newFunction() string",
			ShouldNotContain:  "old implementation",
		})
	})

	t.Run("ReplaceType_ShouldUpdateTypeAndReturnDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-type-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_type_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "type",
			"part_name":     "Config",
			"new_content":   UpdatedType,
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error replacing type",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedSuccess:   true,
			ExpectedFilePath:  testFile.Filepath,
			ExpectedLanguage:  "go",
			ExpectedPartType:  "type",
			ExpectedPartName:  "Config",
			ShouldUpdateFile:  true,
			ShouldContainText: "Version  string",
		})
	})

	t.Run("ReplaceConst_ShouldUpdateConstantAndReturnDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-const-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_const_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "const",
			"part_name":     "ServerPort",
			"new_content":   UpdatedConst,
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error replacing const",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedSuccess:   true,
			ExpectedFilePath:  testFile.Filepath,
			ExpectedLanguage:  "go",
			ExpectedPartType:  "const",
			ExpectedPartName:  "ServerPort",
			ShouldUpdateFile:  true,
			ShouldContainText: `ServerPort = "9000"`,
		})
	})

	t.Run("ReplaceVar_ShouldUpdateVariableAndReturnDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-var-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_var_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "var",
			"part_name":     "GlobalVar",
			"new_content":   UpdatedVar,
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error replacing var",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedSuccess:   true,
			ExpectedFilePath:  testFile.Filepath,
			ExpectedLanguage:  "go",
			ExpectedPartType:  "var",
			ExpectedPartName:  "GlobalVar",
			ShouldUpdateFile:  true,
			ShouldContainText: `GlobalVar = "updated value"`,
		})
	})

	t.Run("ReplaceInterface_ShouldUpdateInterfaceAndReturnDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-interface-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_interface_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "type",
			"part_name":     "UserService",
			"new_content":   UpdatedInterface,
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error replacing interface",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedSuccess:   true,
			ExpectedFilePath:  testFile.Filepath,
			ExpectedLanguage:  "go",
			ExpectedPartType:  "type",
			ExpectedPartName:  "UserService",
			ShouldUpdateFile:  true,
			ShouldContainText: "UpdateUser(user *User) error",
		})
	})

	t.Run("ReplaceMethod_ShouldUpdateMethodAndReturnDetails", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-method-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_method_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "*Config.GetPort",
			"new_content":   UpdatedMethod,
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error replacing method",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedSuccess:   true,
			ExpectedFilePath:  testFile.Filepath,
			ExpectedLanguage:  "go",
			ExpectedPartType:  "func",
			ExpectedPartName:  "*Config.GetPort",
			ShouldUpdateFile:  true,
			ShouldContainText: `if c.Port == ""`,
		})
	})

	t.Run("UnsupportedLanguage_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("unsupported-lang-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("unsupported.py", FileFixtureArgs{
			Content:     "def hello():\n    print('hello')",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "python",
			"part_type":     "func",
			"part_name":     "hello",
			"new_content":   "def hello():\n    print('updated')",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle unsupported language",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "not supported",
		})
	})

	t.Run("InvalidPartType_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("invalid-part-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("invalid_part_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "invalid",
			"part_name":     "something",
			"new_content":   "something",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle invalid part type",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "not supported",
		})
	})

	t.Run("PartNotFound_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("not-found-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("not_found_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "nonexistentFunction",
			"new_content":   "func nonexistentFunction() {}",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle non-existent function",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "not found",
		})
	})
}

func TestValidateFilesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("validate_files")
	require.NotNil(t, tool, "validate_files tool should be registered")

	t.Run("ValidateValidGoFile_ShouldReturnSuccessResult", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("validate-valid-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("valid_test.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"files":         []any{testFile.Filepath},
			"paths":         []any{tf.TempDir()},
			"language":      "go",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error validating valid Go file",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedTotalFiles:   1,
			ExpectedValidFiles:   1,
			ExpectedInvalidFiles: 0,
			ExpectedOverallValid: true,
			ExpectedValidation:   true,
		})
	})

	t.Run("ValidateInvalidGoFile_ShouldReturnValidationErrors", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("validate-invalid-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("invalid_test.go", FileFixtureArgs{
			Content:     "package main\n\nfunc main() {\n    invalid syntax here",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"files":         []any{testFile.Filepath},
			"paths":         []any{tf.TempDir()},
			"language":      "go",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error when validating invalid file",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedTotalFiles:    1,
			ExpectedValidFiles:    0,
			ExpectedInvalidFiles:  1,
			ExpectedOverallValid:  false,
			ExpectedValidation:    false,
			CheckValidationErrors: true,
		})
	})

	t.Run("ValidateDirectory_ShouldProcessMultipleFiles", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("validate-dir-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf.AddFileFixture("valid.go", FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})
		pf.AddFileFixture("invalid.go", FileFixtureArgs{
			Content:     "package main\n\nfunc main() { invalid",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"files":         []any{},
			"paths":         []any{tf.TempDir()},
			"language":      "go",
			"extensions":    []any{".go"},
			"recursive":     true,
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error validating directory",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedTotalFiles:    2,
			ExpectedValidFiles:    1,
			ExpectedInvalidFiles:  1,
			ExpectedOverallValid:  false,
			CheckValidationErrors: true,
		})
	})

	t.Run("UnsupportedLanguage_ShouldHandleGracefully", func(t *testing.T) {
		tf := NewTestFixture(LanguageAwareDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("unsupported-validate-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("test.py", FileFixtureArgs{
			Content:     "print('hello')",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"files":         []any{testFile.Filepath},
			"paths":         []any{tf.TempDir()},
			"language":      "python",
		})

		result, err := getToolResult[LanguageAwareResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should handle unsupported language gracefully",
		)

		requireLanguageAwareResult(t, result, err, LanguageAwareResultOpts{
			ExpectedTotalFiles:    1,
			ExpectedValidFiles:    0,
			ExpectedInvalidFiles:  1,
			ExpectedOverallValid:  false,
			CheckValidationErrors: true,
		})
	})
}
