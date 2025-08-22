package mcptools_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const GetConfigDirPrefix = "get-config-tool-test"

// Get config tool result type
type ConfigResult struct {
	ServerName     string   `json:"server_name"`
	AllowedPaths   []string `json:"allowed_paths"`
	AllowedOrigins []string `json:"allowed_origins"`
	PathCount      int      `json:"path_count"`
	ConfigFilePath string   `json:"config_file_path"`
	HomeDirectory  string   `json:"home_directory"`
	ServerPort     string   `json:"server_port"`
	Summary        string   `json:"summary"`
}

type configToolResultOpts struct {
	ExpectError        bool
	ExpectedErrorMsg   string
	ExpectMinPaths     int
	ExpectedServerName string
}

func requireConfigResult(t *testing.T, result *ConfigResult, err error, opts configToolResultOpts) {
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

	// Check basic fields
	assert.NotEmpty(t, result.ServerName, "Server name should not be empty")
	if opts.ExpectedServerName != "" {
		assert.Equal(t, opts.ExpectedServerName, result.ServerName, "Server name should match expected")
	}

	assert.NotEmpty(t, result.HomeDirectory, "Home directory should not be empty")
	assert.NotEmpty(t, result.Summary, "Summary should not be empty")
	assert.Greater(t, result.PathCount, 0, "Should have at least one allowed path")

	if opts.ExpectMinPaths > 0 {
		assert.GreaterOrEqual(t, len(result.AllowedPaths), opts.ExpectMinPaths, "Should have minimum paths")
		assert.GreaterOrEqual(t, result.PathCount, opts.ExpectMinPaths, "Path count should match minimum")
	}
}

func TestGetConfigTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("get_config")
	require.NotNil(t, tool, "get_config tool should be registered")

	t.Run("GetBasicConfig", func(t *testing.T) {
		tf := fsfix.NewRootFixture(GetConfigDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{
			"session_token": testToken,
		})

		result, err := mcputil.GetToolResult[ConfigResult](mcputil.CallResult(mcputil.CallTool(tool, req)), "Should not error getting config")

		requireConfigResult(t, result, err, configToolResultOpts{
			ExpectMinPaths: 1,
		})

		// Verify the test directory is in allowed paths
		assert.Contains(t, result.AllowedPaths, tf.TempDir(), "Test directory should be in allowed paths")
	})
}
