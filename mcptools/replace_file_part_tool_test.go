package mcptools_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ReplaceFilePartDirPrefix = "replace-file-part-tool-test"

const (
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

// Replace file part tool result type
type ReplaceFilePartResult struct {
	Success     bool   `json:"success"`
	FilePath    string `json:"file_path"`
	Language    string `json:"language"`
	PartType    string `json:"part_type"`
	PartName    string `json:"part_name"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	Message     string `json:"message"`
}

type replaceFilePartResultOpts struct {
	ExpectError       bool
	ExpectedErrorMsg  string
	ExpectedSuccess   bool
	ExpectedFilePath  string
	ExpectedLanguage  string
	ExpectedPartType  string
	ExpectedPartName  string
	ShouldUpdateFile  bool
	ShouldContainText string
	ShouldNotContain  string
	ExpectedContent   string
}

func requireReplaceFilePartResult(t *testing.T, result *ReplaceFilePartResult, err error, opts replaceFilePartResultOpts) {
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

	if opts.ExpectedSuccess {
		assert.True(t, result.Success, "Operation should be successful")
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

func TestReplaceFilePartTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("replace_file_part")
	require.NotNil(t, tool, "replace_file_part tool should be registered")

	t.Run("ReplaceFunction_ShouldUpdateFunctionAndReturnDetails", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReplaceFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-func-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_func_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "oldFunction",
			"new_content":   UpdatedFunction,
		})

		result, err := mcputil.GetToolResult[ReplaceFilePartResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error replacing function")

		requireReplaceFilePartResult(t, result, err, replaceFilePartResultOpts{
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
		tf := testutil.NewTestFixture(ReplaceFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-type-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_type_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "type",
			"part_name":     "Config",
			"new_content":   UpdatedType,
		})

		result, err := mcputil.GetToolResult[ReplaceFilePartResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error replacing type")

		requireReplaceFilePartResult(t, result, err, replaceFilePartResultOpts{
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
		tf := testutil.NewTestFixture(ReplaceFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-const-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_const_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "const",
			"part_name":     "ServerPort",
			"new_content":   UpdatedConst,
		})

		result, err := mcputil.GetToolResult[ReplaceFilePartResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error replacing const")

		requireReplaceFilePartResult(t, result, err, replaceFilePartResultOpts{
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
		tf := testutil.NewTestFixture(ReplaceFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-var-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_var_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "var",
			"part_name":     "GlobalVar",
			"new_content":   UpdatedVar,
		})

		result, err := mcputil.GetToolResult[ReplaceFilePartResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error replacing var")

		requireReplaceFilePartResult(t, result, err, replaceFilePartResultOpts{
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
		tf := testutil.NewTestFixture(ReplaceFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-interface-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_interface_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "type",
			"part_name":     "UserService",
			"new_content":   UpdatedInterface,
		})

		result, err := mcputil.GetToolResult[ReplaceFilePartResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error replacing interface")

		requireReplaceFilePartResult(t, result, err, replaceFilePartResultOpts{
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
		tf := testutil.NewTestFixture(ReplaceFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("replace-method-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("replace_method_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "*Config.GetPort",
			"new_content":   UpdatedMethod,
		})

		result, err := mcputil.GetToolResult[ReplaceFilePartResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should not error replacing method")

		requireReplaceFilePartResult(t, result, err, replaceFilePartResultOpts{
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
		tf := testutil.NewTestFixture(ReplaceFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("unsupported-lang-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("unsupported.py", testutil.FileFixtureArgs{
			Content:     "def hello():\n    print('hello')",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "python",
			"part_type":     "func",
			"part_name":     "hello",
			"new_content":   "def hello():\n    print('updated')",
		})

		result, err := mcputil.GetToolResult[ReplaceFilePartResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should handle unsupported language")

		requireReplaceFilePartResult(t, result, err, replaceFilePartResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "not supported",
		})
	})

	t.Run("InvalidPartType_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReplaceFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("invalid-part-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("invalid_part_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "invalid",
			"part_name":     "something",
			"new_content":   "something",
		})

		result, err := mcputil.GetToolResult[ReplaceFilePartResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should handle invalid part type")

		requireReplaceFilePartResult(t, result, err, replaceFilePartResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "not supported",
		})
	})

	t.Run("PartNotFound_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := testutil.NewTestFixture(ReplaceFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("not-found-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("not_found_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "nonexistentFunction",
			"new_content":   "func nonexistentFunction() {}",
		})

		result, err := mcputil.GetToolResult[ReplaceFilePartResult](mcputil.CallResult(testutil.CallTool(tool, req)), "Should handle non-existent function")

		requireReplaceFilePartResult(t, result, err, replaceFilePartResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "not found",
		})
	})
}
