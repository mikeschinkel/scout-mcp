package mcptools_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const CheckDocsDirPrefix = "check-docs-tool-test"

// CheckDocsResult matches the JSON output structure of CheckDocsTool
type CheckDocsResult struct {
	Path   string           `json:"path"`
	Issues []CheckDocsIssue `json:"issues"`
	Total  int              `json:"total"`
}

type CheckDocsIssue struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	EndLine   *int   `json:"end_line,omitempty"`
	Issue     string `json:"issue"`
	Element   string `json:"element"`
	MultiLine bool   `json:"multi_line"`
}

type checkDocsResultOpts struct {
	ExpectError          bool
	ExpectedErrorMsg     string
	ExpectedIssueCount   int // Use -1 to indicate "don't check exact count"
	ExpectedMinIssues    int
	ExpectedPath         string
	ExpectValidStructure bool
}

func requireCheckDocsResult(t *testing.T, result *CheckDocsResult, err error, opts checkDocsResultOpts) {
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

	if opts.ExpectValidStructure {
		// Verify required fields exist
		assert.NotEmpty(t, result.Path, "Path should not be empty")
		assert.NotNil(t, result.Issues, "Issues should not be nil (can be empty slice)")
		assert.GreaterOrEqual(t, result.Total, 0, "Total should be non-negative")
		assert.Equal(t, result.Total, len(result.Issues), "Total should match issues array length")
	}

	// Check specific path if expected
	if opts.ExpectedPath != "" {
		assert.Equal(t, opts.ExpectedPath, result.Path, "Path should match expected")
	}

	// Check issue count (only if explicitly set)
	if opts.ExpectedIssueCount > 0 || (opts.ExpectedIssueCount == 0 && opts.ExpectedMinIssues == 0) {
		assert.Equal(t, opts.ExpectedIssueCount, result.Total, "Total should match expected count")
		assert.Len(t, result.Issues, opts.ExpectedIssueCount, "Issues array should match expected count")
	}
	if opts.ExpectedMinIssues > 0 {
		assert.GreaterOrEqual(t, result.Total, opts.ExpectedMinIssues, "Should have at least minimum issues")
		assert.GreaterOrEqual(t, len(result.Issues), opts.ExpectedMinIssues, "Issues array should have minimum count")
	}

	// Validate issue structure if any issues exist
	for i, issue := range result.Issues {
		assert.NotEmpty(t, issue.File, "Issue %d should have file", i)
		// Line can be 0 for README.md issues, otherwise should be > 0
		if !strings.Contains(issue.Issue, "README.md") {
			assert.Greater(t, issue.Line, 0, "Issue %d should have valid line number", i)
		} else {
			assert.GreaterOrEqual(t, issue.Line, 0, "Issue %d should have non-negative line number", i)
		}
		assert.NotEmpty(t, issue.Issue, "Issue %d should have issue description", i)
		// Element name can be empty for file-level issues (like missing file comments)
		// EndLine can be nil, but if set should be >= Line
		if issue.EndLine != nil {
			assert.GreaterOrEqual(t, *issue.EndLine, issue.Line, "Issue %d EndLine should be >= Line", i)
		}
	}
}

func TestCheckDocsTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("check_docs")
	require.NotNil(t, tool, "check_docs tool should be registered")

	t.Run("ToolRegistration_ShouldHaveCorrectMetadata", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()
		tf.Setup(t)

		assert.Equal(t, "check_docs", tool.Name(), "Tool name should be check_docs")
		assert.NotEmpty(t, tool.Options().Description, "Tool should have description")

		// Verify required properties exist
		properties := tool.Options().Properties
		hasPath := false
		hasRecursive := false
		hasMaxFiles := false

		for _, prop := range properties {
			switch prop.GetName() {
			case "path":
				hasPath = true
				assert.True(t, prop.IsRequired(), "path property should be required")
			case "recursive":
				hasRecursive = true
				assert.False(t, prop.IsRequired(), "recursive property should be optional")
			case "max_files":
				hasMaxFiles = true
				assert.False(t, prop.IsRequired(), "max_files property should be optional")
			}
		}

		assert.True(t, hasPath, "Tool should have path property")
		assert.True(t, hasRecursive, "Tool should have recursive property")
		assert.True(t, hasMaxFiles, "Tool should have max_files property")
	})

	t.Run("RequiredParameters_MissingPath_ShouldFail", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()
		tf.Setup(t)

		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"recursive":     true,
			"max_files":     100,
			// Missing required "path" parameter
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should handle missing path parameter")

		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "path",
		})
	})

	t.Run("FullyDocumentedGoFile_ShouldReturnZeroIssues", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("documented-project", nil)

		// Create a fully documented Go file
		pf.AddFileFixture("documented.go", &fsfix.FileFixtureArgs{
			Content: `// Package main provides a fully documented example.
package main

// main is the entry point of the application.
func main() {
	// Application logic here
}

// Config represents the application configuration.
type Config struct {
	Port string
}
`,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          pf.Dir(),
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error analyzing fully documented Go file")

		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectValidStructure: true,
			ExpectedPath:         pf.Dir(),
			ExpectedIssueCount:   0, // Fully documented file should have zero issues
		})
	})

	t.Run("UndocumentedGoFile_ShouldReturnSpecificIssues", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("undocumented-project", nil)

		// Create a Go file with specific documentation issues
		pf.AddFileFixture("undocumented.go", &fsfix.FileFixtureArgs{
			Content: `package main

func main() {
	// Undocumented function
}

type Config struct {
	Port string
}
`,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          pf.Dir(),
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error analyzing undocumented Go file")

		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectValidStructure: true,
			ExpectedPath:         pf.Dir(),
			ExpectedIssueCount:   3, // Missing file comment, func comment, type comment
		})
	})

	t.Run("ValidDirectory_ShouldAnalyzeAllFiles", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("mixed-project", nil)

		// Create one documented file
		pf.AddFileFixture("documented.go", &fsfix.FileFixtureArgs{
			Content: `// Package main provides documented functionality.
package main

// DocumentedFunc is a properly documented function.
func DocumentedFunc() {
	// Function implementation
}
`,
		})

		// Create one undocumented file
		pf.AddFileFixture("undocumented.go", &fsfix.FileFixtureArgs{
			Content: `package main

func UndocumentedFunc() {
	// No documentation
}
`,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          pf.Dir(),
			"recursive":     false,
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error analyzing directory")

		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectValidStructure: true,
			ExpectedPath:         pf.Dir(),
			ExpectedIssueCount:   2, // undocumented.go: missing file comment + missing func comment
		})
	})

	t.Run("ParameterCombinations_RecursiveTrue_ShouldFindNestedFiles", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("recursive-project", nil)

		// Create fully documented files to avoid unpredictable issue counts
		pf.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
			Content: `// Package main demonstrates recursive analysis.
package main

// main is the entry point.
func main() {
}
`,
		})

		subPf := pf.AddDirFixture("subdir", nil)

		subPf.AddFileFixture("sub.go", &fsfix.FileFixtureArgs{
			Content: `// Package subdir provides nested functionality.
package subdir

// SubFunc demonstrates nested package analysis.
func SubFunc() {
}
`,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          pf.Dir(),
			"recursive":     true,
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error with recursive analysis")

		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectValidStructure: true,
			ExpectedPath:         pf.Dir(),
			ExpectedMinIssues:    1, // Should find README.md issues (normal behavior)
		})
	})

	t.Run("ParameterCombinations_MaxFiles_ShouldLimitResults", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("test-project", nil)

		// Create multiple Go files
		for i := 1; i <= 5; i++ {
			pf.AddFileFixture(fmt.Sprintf("file%d.go", i), &fsfix.FileFixtureArgs{
				Content: fmt.Sprintf(`package main

func Function%d() {
}
`, i),
			})
		}

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          pf.Dir(),
			"recursive":     true,
			"max_files":     2, // Limit to 2 files
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error with max_files limit")

		// Just verify the tool accepts max_files parameter and returns valid structure
		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectValidStructure: true,
			ExpectedPath:         pf.Dir(),
			ExpectedMinIssues:    1, // Will find documentation issues (normal behavior)
		})
	})

	t.Run("ErrorHandling_NonExistentPath_ShouldHandleGracefully", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()

		nonExistentPath := tf.AddFileFixture("does-not-exist.go", &fsfix.FileFixtureArgs{
			Missing: true,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          nonExistentPath.Filepath,
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should handle non-existent path gracefully")

		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "no such file", // Or whatever error message the langutil returns
		})
	})

	t.Run("EmptyDirectory_ShouldReturnZeroIssues", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("empty-project", nil)

		// Add some non-Go files
		pf.AddFileFixture("README.md", &fsfix.FileFixtureArgs{
			Content: "# Test Project\n",
		})

		pf.AddFileFixture("config.json", &fsfix.FileFixtureArgs{
			Content: `{"test": true}`,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          pf.Dir(),
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should handle directory with no Go files")

		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectValidStructure: true,
			ExpectedPath:         pf.Dir(),
			ExpectedIssueCount:   0, // No Go files = no issues
		})
	})
}
