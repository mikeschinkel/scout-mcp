package mcptools_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ValidateFilesDirPrefix = "validate-files-tool-test"

// Validate files tool result types
type ValidateFilesResult struct {
	TotalFiles   int                `json:"total_files"`
	ValidFiles   int                `json:"valid_files"`
	InvalidFiles int                `json:"invalid_files"`
	OverallValid bool               `json:"overall_valid"`
	Results      []ValidationResult `json:"results"`
}

type ValidationResult struct {
	FilePath string `json:"file_path"`
	Language string `json:"language"`
	Valid    bool   `json:"valid"`
	Error    string `json:"error,omitempty"`
}

type validateFilesResultOpts struct {
	ExpectError           bool
	ExpectedErrorMsg      string
	ExpectedTotalFiles    int
	ExpectedValidFiles    int
	ExpectedInvalidFiles  int
	ExpectedOverallValid  bool
	ExpectedValidation    bool
	CheckValidationErrors bool
}

func requireValidateFilesResult(t *testing.T, result *ValidateFilesResult, err error, opts validateFilesResultOpts) {
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
}

func TestValidateFilesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("validate_files")
	require.NotNil(t, tool, "validate_files tool should be registered")

	t.Run("ValidateValidGoFile_ShouldReturnSuccessResult", func(t *testing.T) {
		tf := fsfix.NewRootFixture(ValidateFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("validate-valid-project", nil)
		testFile := pf.AddFileFixture("valid_test.go", &fsfix.FileFixtureArgs{
			Content: GoTestContent,
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"files":         []any{testFile.Filepath},
			//"paths":         []any{tf.TempDir()},
			"language": "go",
		})

		result, err := mcputil.GetToolResult[ValidateFilesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error validating valid Go file")

		requireValidateFilesResult(t, result, err, validateFilesResultOpts{
			ExpectedTotalFiles:   1,
			ExpectedValidFiles:   1,
			ExpectedInvalidFiles: 0,
			ExpectedOverallValid: true,
			ExpectedValidation:   true,
		})
	})

	t.Run("ValidateInvalidGoFile_ShouldReturnValidationErrors", func(t *testing.T) {
		tf := fsfix.NewRootFixture(ValidateFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("validate-invalid-project", nil)
		testFile := pf.AddFileFixture("invalid_test.go", &fsfix.FileFixtureArgs{
			Content: "package main\n\nfunc main() {\n    invalid syntax here",
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"files":         []any{testFile.Filepath},
			//"paths":         []any{tf.TempDir()},
			"language": "go",
		})

		result, err := mcputil.GetToolResult[ValidateFilesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error when validating invalid file")

		requireValidateFilesResult(t, result, err, validateFilesResultOpts{
			ExpectedTotalFiles:    1,
			ExpectedValidFiles:    0,
			ExpectedInvalidFiles:  1,
			ExpectedOverallValid:  false,
			ExpectedValidation:    false,
			CheckValidationErrors: true,
		})
	})

	t.Run("ValidateDirectory_ShouldProcessMultipleFiles", func(t *testing.T) {
		tf := fsfix.NewRootFixture(ValidateFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("validate-dir-project", nil)
		pf.AddFileFixture("valid.go", &fsfix.FileFixtureArgs{
			Content: GoTestContent,
		})
		pf.AddFileFixture("invalid.go", &fsfix.FileFixtureArgs{
			Content: "package main\n\nfunc main() { invalid",
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"files":         []any{},
			"paths":         []any{tf.TempDir()},
			"language":      "go",
			"extensions":    []any{".go"},
			"recursive":     true,
		})

		result, err := mcputil.GetToolResult[ValidateFilesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error validating directory")

		requireValidateFilesResult(t, result, err, validateFilesResultOpts{
			ExpectedTotalFiles:    2,
			ExpectedValidFiles:    1,
			ExpectedInvalidFiles:  1,
			ExpectedOverallValid:  false,
			CheckValidationErrors: true,
		})
	})

	t.Run("UnsupportedLanguage_ShouldHandleGracefully", func(t *testing.T) {
		tf := fsfix.NewRootFixture(ValidateFilesDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("unsupported-validate-project", nil)
		testFile := pf.AddFileFixture("test.py", &fsfix.FileFixtureArgs{
			Content: "print('hello')",
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
			"files":         []any{testFile.Filepath},
			//"paths":         []any{tf.TempDir()},
			"language": "python",
		})

		result, err := mcputil.GetToolResult[ValidateFilesResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should handle unsupported language gracefully")

		requireValidateFilesResult(t, result, err, validateFilesResultOpts{
			ExpectedTotalFiles:    1,
			ExpectedValidFiles:    0,
			ExpectedInvalidFiles:  1,
			ExpectedOverallValid:  false,
			CheckValidationErrors: true,
		})
	})
}
