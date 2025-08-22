package mcputil_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mikeschinkel/scout-mcp/fsfix"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const StartSessionDirPrefix = "start-session-tool-test"

type startSessionResultOpts struct {
	ExpectError           bool
	ExpectedErrorMsg      string
	ShouldHaveToken       bool
	ShouldHaveMessage     bool
	TokenExpiresIn24Hours bool
}

func requireStartSessionResult(t *testing.T, result *mcputil.StartSessionResult, err error, opts startSessionResultOpts) {
	t.Helper()

	if errors.Is(err, mcputil.ErrNoPayloadType) {
		err = nil
	}

	if opts.ExpectError {
		require.Error(t, err, "Should have error")
		if opts.ExpectedErrorMsg != "" {
			assert.Contains(t, err.Error(), opts.ExpectedErrorMsg, "Error should contain expected message")
		}
		return
	}

	require.NoError(t, err, "Should not have error")
	require.NotNil(t, result, "Result should not be nil")

	if opts.ShouldHaveMessage {
		assert.NotEmpty(t, result.Message, "Message should not be empty")
		assert.Contains(t, result.Message, "MCP Session Started", "Message should indicate session started")
	}

	if opts.ShouldHaveToken {
		assert.NotEmpty(t, result.SessionToken, "Session token should not be empty")
		assert.False(t, result.TokenExpiresAt.IsZero(), "Token expiration should be set")
		assert.True(t, result.TokenExpiresAt.After(time.Now()), "Token should expire in the future")
	}

	if opts.TokenExpiresIn24Hours {
		// Verify token expires within 24 hours (with some tolerance)
		now := time.Now()
		expectedExpiry := now.Add(24 * time.Hour)
		tolerance := 5 * time.Minute

		assert.True(t, result.TokenExpiresAt.After(now), "Token should expire in the future")
		assert.True(t, result.TokenExpiresAt.Before(expectedExpiry.Add(tolerance)), "Token should expire within 24 hours + tolerance")
		assert.True(t, result.TokenExpiresAt.After(expectedExpiry.Add(-tolerance)), "Token should expire within 24 hours - tolerance")
	}

}

func TestStartSessionTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("start_session")
	require.NotNil(t, tool, "start_session tool should be registered")

	t.Run("BasicSessionCreation_ShouldReturnValidSessionData", func(t *testing.T) {
		tf := fsfix.NewRootFixture(StartSessionDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddRepoFixture("test-project", nil)
		// Add a test file to make it detectable as a project
		pf.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
			Content: "package main\n\nfunc main() {}\n",
		})

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		// start_session tool doesn't require any parameters
		req := mcputil.NewMockRequest(mcputil.Params{})

		result, err := mcputil.GetToolResult[mcputil.StartSessionResult](
			mcputil.CallResult(mcputil.CallTool(tool, req)),
			"Should not error creating session",
		)
		requireStartSessionResult(t, result, err, startSessionResultOpts{
			ShouldHaveToken:   true,
			ShouldHaveMessage: true,
		})

		// Verify session token is properly formatted
		assert.Len(t, result.SessionToken, 64, "Session token should be 64 characters (32 bytes hex)")

	})

	t.Run("NoAllowedPaths_ShouldStillCreateSession", func(t *testing.T) {
		tf := fsfix.NewRootFixture(StartSessionDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{}, // No allowed paths
		}))

		req := mcputil.NewMockRequest(mcputil.Params{})

		result, err := mcputil.GetToolResult[mcputil.StartSessionResult](
			mcputil.CallResult(mcputil.CallTool(tool, req)),
			"Should not error creating session without allowed paths",
		)

		requireStartSessionResult(t, result, err, startSessionResultOpts{
			ShouldHaveToken:   true,
			ShouldHaveMessage: true,
		})

	})

	t.Run("TokenExpiration_ShouldBeWithin24Hours", func(t *testing.T) {
		tf := fsfix.NewRootFixture(StartSessionDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(mcputil.NewMockConfig(mcputil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := mcputil.NewMockRequest(mcputil.Params{})

		result, err := mcputil.GetToolResult[mcputil.StartSessionResult](
			mcputil.CallResult(mcputil.CallTool(tool, req)),
			"Should not error creating session",
		)

		requireStartSessionResult(t, result, err, startSessionResultOpts{
			ShouldHaveToken:       true,
			TokenExpiresIn24Hours: true,
		})

	})
}
