package mcptools_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const FindFilePartDirPrefix = "find-file-part-tool-test"

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
)

// Find file part tool result type
type FindFilePartResult struct {
	Found       bool   `json:"found"`
	FilePath    string `json:"file_path"`
	Language    string `json:"language"`
	PartType    string `json:"part_type"`
	PartName    string `json:"part_name"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	Content     string `json:"content"`
}

type findFilePartResultOpts struct {
	ExpectError         bool
	ExpectedErrorMsg    string
	ExpectedFound       bool
	ExpectedFilePath    string
	ExpectedLanguage    string
	ExpectedPartType    string
	ExpectedPartName    string
	ExpectedStartLine   int
	ExpectedEndLine     int
	ExpectedStartOffset int
	ExpectedEndOffset   int
}

func requireFindFilePartResult(t *testing.T, result *FindFilePartResult, err error, opts findFilePartResultOpts) {
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
}

func TestFindFilePartTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("find_file_part")
	require.NotNil(t, tool, "find_file_part tool should be registered")

	t.Run("FindGoFunction_ShouldLocateAndReturnFunctionDetails", func(t *testing.T) {
		tf := testutil.NewTestFixture(FindFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("find-func-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "oldFunction",
		})

		result, err := mcputil.GetToolResult[FindFilePartResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error finding function")

		requireFindFilePartResult(t, result, err, findFilePartResultOpts{
			ExpectedFound:    true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedPartType: "func",
			ExpectedPartName: "oldFunction",
		})
	})

	t.Run("FindGoType_ShouldLocateAndReturnTypeDetails", func(t *testing.T) {
		tf := testutil.NewTestFixture(FindFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("find-type-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_type_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "type",
			"part_name":     "Config",
		})

		result, err := mcputil.GetToolResult[FindFilePartResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error finding type")

		requireFindFilePartResult(t, result, err, findFilePartResultOpts{
			ExpectedFound:    true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedPartType: "type",
			ExpectedPartName: "Config",
		})
	})

	t.Run("FindGoConst_ShouldLocateAndReturnConstDetails", func(t *testing.T) {
		tf := testutil.NewTestFixture(FindFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("find-const-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_const_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "const",
			"part_name":     "ServerPort",
		})

		result, err := mcputil.GetToolResult[FindFilePartResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error finding const")

		requireFindFilePartResult(t, result, err, findFilePartResultOpts{
			ExpectedFound:    true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedPartType: "const",
			ExpectedPartName: "ServerPort",
		})
	})

	t.Run("FindGoVar_ShouldLocateAndReturnVarDetails", func(t *testing.T) {
		tf := testutil.NewTestFixture(FindFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("find-var-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_var_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "var",
			"part_name":     "GlobalVar",
		})

		result, err := mcputil.GetToolResult[FindFilePartResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error finding var")

		requireFindFilePartResult(t, result, err, findFilePartResultOpts{
			ExpectedFound:    true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedPartType: "var",
			ExpectedPartName: "GlobalVar",
		})
	})

	t.Run("FindMethod_ShouldLocateAndReturnMethodDetails", func(t *testing.T) {
		tf := testutil.NewTestFixture(FindFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("find-method-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_method_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "*Config.GetPort",
		})

		result, err := mcputil.GetToolResult[FindFilePartResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error finding method")

		requireFindFilePartResult(t, result, err, findFilePartResultOpts{
			ExpectedFound:    true,
			ExpectedFilePath: testFile.Filepath,
			ExpectedPartType: "func",
			ExpectedPartName: "*Config.GetPort",
		})
	})

	t.Run("PartNotFound_ShouldReturnErrorWithMessage", func(t *testing.T) {
		tf := testutil.NewTestFixture(FindFilePartDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("missing-part-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		testFile := pf.AddFileFixture("find_missing_test.go", testutil.FileFixtureArgs{
			Content:     GoTestContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          testFile.Filepath,
			"language":      "go",
			"part_type":     "func",
			"part_name":     "nonexistentFunction",
		})

		result, err := mcputil.GetToolResult[FindFilePartResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should handle part not found")

		requireFindFilePartResult(t, result, err, findFilePartResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "not found",
		})
	})
}
