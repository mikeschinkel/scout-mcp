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
	Path           string           `json:"path"`
	Issues         []CheckDocsIssue `json:"issues"`
	ReturnedCount  int              `json:"returned_count"`
	TotalCount     int              `json:"total_count"`
	RemainingCount int              `json:"remaining_count"`
	SizeLimited    bool             `json:"size_limited"`
	ResponseSize   int              `json:"response_size_chars"`
	Message        string           `json:"message,omitempty"`
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
	ExpectError           bool
	ExpectedErrorMsg      string
	ExpectedIssueCount    int // Use -1 to indicate "don't check exact count" - for tests where TotalCount == Issues.length
	ExpectedTotalCount    int // Use -1 to indicate "don't check" - checks TotalCount specifically
	ExpectedReturnedCount int // Use -1 to indicate "don't check" - checks Issues.length specifically
	ExpectedMinIssues     int
	ExpectedPath          string
	ExpectValidStructure  bool
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
		assert.GreaterOrEqual(t, result.TotalCount, 0, "TotalCount should be non-negative")
		assert.Equal(t, result.ReturnedCount, len(result.Issues), "ReturnedCount should match issues array length")

		// Only validate RemainingCount with old logic if not using offset-specific validation
		if opts.ExpectedTotalCount == 0 && opts.ExpectedReturnedCount == 0 {
			assert.Equal(t, result.TotalCount-result.ReturnedCount, result.RemainingCount, "RemainingCount should be correct")
		}
	}

	// Check specific path if expected
	if opts.ExpectedPath != "" {
		assert.Equal(t, opts.ExpectedPath, result.Path, "Path should match expected")
	}

	// Check issue count (only if explicitly set) - this is for tests where TotalCount == Issues.length
	if opts.ExpectedIssueCount > 0 || (opts.ExpectedIssueCount == 0 && opts.ExpectedMinIssues == 0) {
		assert.Equal(t, opts.ExpectedIssueCount, result.TotalCount, "TotalCount should match expected count")
		assert.Len(t, result.Issues, opts.ExpectedIssueCount, "Issues array should match expected count")
	}

	// Check total count specifically (for offset/limit scenarios) - only check if ExpectedIssueCount disabled
	if opts.ExpectedTotalCount > 0 && opts.ExpectedIssueCount == -1 {
		assert.Equal(t, opts.ExpectedTotalCount, result.TotalCount, "TotalCount should match expected total count")
	}

	// Check returned count specifically (for offset/limit scenarios) - only check if ExpectedIssueCount disabled
	if opts.ExpectedReturnedCount >= 0 && opts.ExpectedIssueCount == -1 {
		assert.Len(t, result.Issues, opts.ExpectedReturnedCount, "Issues array should match expected returned count")
	}

	if opts.ExpectedMinIssues > 0 {
		assert.GreaterOrEqual(t, result.TotalCount, opts.ExpectedMinIssues, "Should have at least minimum issues")
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
		hasLanguage := false
		hasOffset := false

		for _, prop := range properties {
			switch prop.GetName() {
			case "path":
				hasPath = true
				assert.True(t, prop.IsRequired(), "path property should be required")
			case "recursive":
				hasRecursive = true
				assert.False(t, prop.IsRequired(), "recursive property should be optional")
			case "language":
				hasLanguage = true
				assert.True(t, prop.IsRequired(), "language property should be required")
			case "offset":
				hasOffset = true
				assert.False(t, prop.IsRequired(), "offset property should be optional")
			}
		}

		assert.True(t, hasPath, "Tool should have path property")
		assert.True(t, hasRecursive, "Tool should have recursive property")
		assert.True(t, hasLanguage, "Tool should have language property")
		assert.True(t, hasOffset, "Tool should have offset property")
	})

	t.Run("RequiredParameters_MissingLanguage_ShouldFail", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()
		tf.Setup(t)

		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          tf.TempDir(),
			"recursive":     true,
			// Missing required "language" parameter
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should handle missing language parameter")

		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectError:      true,
			ExpectedErrorMsg: "language",
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
			"language":      "go",
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
			"language":      "go",
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
			"language":      "go",
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
			"language":      "go",
			"recursive":     true,
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error with recursive analysis")

		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectValidStructure: true,
			ExpectedPath:         pf.Dir(),
			ExpectedMinIssues:    1, // Should find README.md issues (normal behavior)
		})
	})

	t.Run("ParameterCombinations_Language_ShouldWork", func(t *testing.T) {
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
			"language":      "go",
			"recursive":     true,
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should work with language parameter")

		// Just verify the tool accepts language parameter and returns valid structure
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
			"language":      "go",
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
			"language":      "go",
		})

		result, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should handle directory with no Go files")

		requireCheckDocsResult(t, result, err, checkDocsResultOpts{
			ExpectValidStructure: true,
			ExpectedPath:         pf.Dir(),
			ExpectedIssueCount:   0, // No Go files = no issues
		})
	})

	t.Run("OffsetParameter_ShouldSkipSpecifiedNumberOfIssues", func(t *testing.T) {
		tf := fsfix.NewRootFixture(CheckDocsDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("offset-test-project", nil)

		// Create multiple Go files with predictable documentation issues
		for i := 1; i <= 5; i++ {
			pf.AddFileFixture(fmt.Sprintf("file%d.go", i), &fsfix.FileFixtureArgs{
				Content: fmt.Sprintf(`package main

func Function%d() {
	// Missing documentation
}

type Type%d struct {
	Field string
}
`, i, i),
			})
		}

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		// First, get all issues without offset
		reqAll := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          pf.Dir(),
			"language":      "go",
		})

		allResult, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, reqAll)), "Should get all issues")
		requireCheckDocsResult(t, allResult, err, checkDocsResultOpts{
			ExpectValidStructure: true,
			ExpectedPath:         pf.Dir(),
			ExpectedMinIssues:    1, // Should have multiple issues
		})

		totalIssues := allResult.TotalCount
		if totalIssues < 3 {
			t.Skipf("Need at least 3 issues for offset testing, got %d", totalIssues)
		}

		// Test offset = 2 (skip first 2 issues)
		expectedReturnedCount := totalIssues - 2
		if expectedReturnedCount < 0 {
			expectedReturnedCount = 0
		}

		reqOffset := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          pf.Dir(),
			"language":      "go",
			"offset":        2,
		})

		offsetResult, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, reqOffset)), "Should handle offset parameter")
		requireCheckDocsResult(t, offsetResult, err, checkDocsResultOpts{
			ExpectValidStructure:  true,
			ExpectedPath:          pf.Dir(),
			ExpectedIssueCount:    -1,                    // Disable original logic
			ExpectedTotalCount:    totalIssues,           // Total should remain the same
			ExpectedReturnedCount: expectedReturnedCount, // Returned should be reduced by offset
		})

		// Verify that RemainingCount is calculated correctly: total - offset - returned
		expectedRemainingCount := totalIssues - 2 - expectedReturnedCount // 15 - 2 - 13 = 0
		assert.Equal(t, expectedRemainingCount, offsetResult.RemainingCount,
			"RemainingCount should be total(%d) - offset(2) - returned(%d) = %d",
			totalIssues, expectedReturnedCount, expectedRemainingCount)

		// Test offset beyond available issues
		reqBeyond := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"path":          pf.Dir(),
			"language":      "go",
			"offset":        totalIssues + 10,
		})

		beyondResult, err := mcputil.GetToolResult[CheckDocsResult](mcputil.CallResult(mcputil.CallTool(tool, reqBeyond)), "Should handle offset beyond available issues")
		requireCheckDocsResult(t, beyondResult, err, checkDocsResultOpts{
			ExpectValidStructure:  true,
			ExpectedPath:          pf.Dir(),
			ExpectedIssueCount:    -1,          // Disable original logic
			ExpectedTotalCount:    totalIssues, // Total should remain the same
			ExpectedReturnedCount: 0,           // Should return 0 issues when offset beyond total
		})

		// Verify that RemainingCount is now correctly 0 (the bug we fixed)
		assert.Equal(t, 0, beyondResult.RemainingCount,
			"RemainingCount should be 0 when offset is beyond total issues")
	})
}
