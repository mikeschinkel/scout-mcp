package mcptools_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const AnalyzeFilesDirPrefix = "analyze-files-tool-test"

type AnalyzeFilesResult struct {
	Files []struct {
		Path  string `json:"path"`
		Name  string `json:"name"`
		Size  int64  `json:"size"`
		Lines int    `json:"lines,omitempty"`
		Error string `json:"error,omitempty"`
	} `json:"files"`
	TotalFiles  int    `json:"total_files"`
	TotalSize   int64  `json:"total_size"`
	TotalErrors int    `json:"total_errors"`
	Summary     string `json:"summary"`
}

type analyzeFilesToolResultOpts struct {
	ExpectError       bool
	ExpectedErrorMsg  string
	ExpectFiles       int
	ExpectMinFiles    int
	ExpectedTotalSize int64
}

func requireAnalyzeFilesResult(t *testing.T, result *AnalyzeFilesResult, err error, opts analyzeFilesToolResultOpts) {
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

	// Check files count
	if opts.ExpectFiles > 0 {
		assert.Equal(t, opts.ExpectFiles, result.TotalFiles, "Should have expected number of files")
		assert.Len(t, result.Files, opts.ExpectFiles, "Files array should match expected count")
	}
	if opts.ExpectMinFiles > 0 {
		assert.GreaterOrEqual(t, result.TotalFiles, opts.ExpectMinFiles, "Should have at least minimum files")
		assert.GreaterOrEqual(t, len(result.Files), opts.ExpectMinFiles, "Files array should have minimum count")
	}

	// Check total size if expected
	if opts.ExpectedTotalSize > 0 {
		assert.Equal(t, opts.ExpectedTotalSize, result.TotalSize, "Total size should match expected")
	}

	assert.NotEmpty(t, result.Summary, "Summary should not be empty")
}

func TestAnalyzeFilesTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("analyze_files")
	require.NotNil(t, tool, "analyze_files tool should be registered")

	t.Run("AnalyzeSingleFile", func(t *testing.T) {
		tf := testutil.NewTestFixture(AnalyzeFilesDirPrefix)
		defer tf.Cleanup()

		// Create a more complex file for analysis
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
		testFile := tf.AddFileFixture("analyze_test.go", testutil.FileFixtureArgs{
			Content:     complexContent,
			Permissions: 0644,
		})
		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"files":         []any{testFile.Filepath},
		})

		result, err := getToolResult[AnalyzeFilesResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error analyzing file",
		)

		requireAnalyzeFilesResult(t, result, err, analyzeFilesToolResultOpts{
			ExpectFiles: 1,
		})
	})

	t.Run("AnalyzeMultipleFiles", func(t *testing.T) {
		tf := testutil.NewTestFixture(AnalyzeFilesDirPrefix)
		defer tf.Cleanup()

		// Create multiple files
		file1 := tf.AddFileFixture("simple.txt", testutil.FileFixtureArgs{
			Content:     "Simple text file\nWith two lines",
			Permissions: 0644,
		})

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
		file2 := tf.AddFileFixture("config.json", testutil.FileFixtureArgs{
			Content:     configContent,
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"files":         []any{file1.Filepath, file2.Filepath},
		})

		result, err := getToolResult[AnalyzeFilesResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error analyzing multiple files",
		)

		requireAnalyzeFilesResult(t, result, err, analyzeFilesToolResultOpts{
			ExpectFiles: 2,
		})
	})

	t.Run("AnalyzeNonExistentFile", func(t *testing.T) {
		tf := testutil.NewTestFixture(AnalyzeFilesDirPrefix)
		defer tf.Cleanup()

		// Add a missing file that doesn't exist
		missingFile := tf.AddFileFixture("nonexistent.txt", testutil.FileFixtureArgs{
			Missing: true,
		})
		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": testToken,
			"files":         []any{missingFile.Filepath},
		})

		result, err := getToolResult[AnalyzeFilesResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with non-existent file",
		)

		requireAnalyzeFilesResult(t, result, err, analyzeFilesToolResultOpts{
			ExpectFiles: 1, // Should report the missing file with error
		})
		// The tool should handle missing files gracefully and report them
		assert.Greater(t, result.TotalErrors, 0, "Should have reported errors for missing file")
	})
}
