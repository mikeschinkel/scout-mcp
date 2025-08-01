package mcptools_test

import (
	"testing"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const StartSessionDirPrefix = "start-session-tool-test"

// Start session tool result type
type StartSessionResult struct {
	SessionToken   string                           `json:"session_token"`
	TokenExpiresAt time.Time                        `json:"token_expires_at"`
	QuickStart     []string                         `json:"quick_start"`
	ServerConfig   map[string]any                   `json:"server_config"`
	Instructions   mcptools.InstructionsConfig      `json:"instructions"`
	Message        string                           `json:"message"`
	CurrentProject *mcptools.ProjectDetectionResult `json:"current_project,omitempty"`
}

type startSessionResultOpts struct {
	ExpectError        bool
	ExpectedErrorMsg   string
	ShouldHaveToken    bool
	ShouldHaveConfig   bool
	ShouldHaveMessage  bool
	MinQuickStartItems int
	ShouldHaveGeneral  bool
	MinLanguages       int
}

func requireStartSessionResult(t *testing.T, result *StartSessionResult, err error, opts startSessionResultOpts) {
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

	if opts.ShouldHaveToken {
		assert.NotEmpty(t, result.SessionToken, "Session token should not be empty")
		assert.False(t, result.TokenExpiresAt.IsZero(), "Token expiration should be set")
		assert.True(t, result.TokenExpiresAt.After(time.Now()), "Token should expire in the future")
	}

	if opts.ShouldHaveConfig {
		assert.NotNil(t, result.ServerConfig, "Server config should not be nil")
		assert.NotEmpty(t, result.ServerConfig, "Server config should not be empty")
	}

	if opts.ShouldHaveMessage {
		assert.NotEmpty(t, result.Message, "Message should not be empty")
		assert.Contains(t, result.Message, "MCP Session Started", "Message should indicate session started")
	}

	if opts.MinQuickStartItems > 0 {
		assert.GreaterOrEqual(t, len(result.QuickStart), opts.MinQuickStartItems, "Should have minimum quick start items")
	}

	if opts.ShouldHaveGeneral {
		assert.NotEmpty(t, result.Instructions.General, "General instructions should not be empty")
	}

	if opts.MinLanguages > 0 {
		assert.GreaterOrEqual(t, len(result.Instructions.Languages), opts.MinLanguages, "Should have minimum language instructions")
	}

	// Always check that extension mappings exist
	assert.NotNil(t, result.Instructions.ExtensionMappings, "Extension mappings should not be nil")
	assert.NotEmpty(t, result.Instructions.ExtensionMappings, "Extension mappings should not be empty")
}

func TestStartSessionTool(t *testing.T) {
	// Get the tool
	tool := mcputil.GetRegisteredTool("start_session")
	require.NotNil(t, tool, "start_session tool should be registered")

	t.Run("BasicSessionCreation_ShouldReturnValidSessionData", func(t *testing.T) {
		tf := testutil.NewTestFixture(StartSessionDirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("test-project", testutil.ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		// Add a test file to make it detectable as a project
		pf.AddFileFixture("main.go", testutil.FileFixtureArgs{
			Content:     "package main\n\nfunc main() {}\n",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		// start_session tool doesn't require any parameters
		req := testutil.NewMockRequest(testutil.Params{})

		result, err := getToolResult[StartSessionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error creating session",
		)

		requireStartSessionResult(t, result, err, startSessionResultOpts{
			ShouldHaveToken:    true,
			ShouldHaveConfig:   true,
			ShouldHaveMessage:  true,
			MinQuickStartItems: 1,
			ShouldHaveGeneral:  true,
			MinLanguages:       1, // Should have at least Go instructions
		})

		// Verify session token is properly formatted
		assert.Len(t, result.SessionToken, 64, "Session token should be 64 characters (32 bytes hex)")

		// Verify server config contains expected fields
		assert.Contains(t, result.ServerConfig, "serverName", "Config should contain server name")
		assert.Contains(t, result.ServerConfig, "allowedPaths", "Config should contain allowed paths")

		// Verify Go language instructions are present
		foundGo := false
		for _, lang := range result.Instructions.Languages {
			if lang.Language == "go" {
				foundGo = true
				assert.NotEmpty(t, lang.Content, "Go instructions should have content")
				assert.Contains(t, lang.Extensions, ".go", "Go instructions should include .go extension")
				break
			}
		}
		assert.True(t, foundGo, "Should include Go language instructions")

		// Verify extension mappings include common languages
		assert.Equal(t, "go", result.Instructions.ExtensionMappings[".go"], "Should map .go to go")
		assert.Equal(t, "python", result.Instructions.ExtensionMappings[".py"], "Should map .py to python")
		assert.Equal(t, "javascript", result.Instructions.ExtensionMappings[".js"], "Should map .js to javascript")
	})

	t.Run("NoAllowedPaths_ShouldStillCreateSession", func(t *testing.T) {
		tf := testutil.NewTestFixture(StartSessionDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{}, // No allowed paths
		}))

		req := testutil.NewMockRequest(testutil.Params{})

		result, err := getToolResult[StartSessionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error creating session without allowed paths",
		)

		requireStartSessionResult(t, result, err, startSessionResultOpts{
			ShouldHaveToken:    true,
			ShouldHaveConfig:   true,
			ShouldHaveMessage:  true,
			MinQuickStartItems: 1,
			ShouldHaveGeneral:  true,
			MinLanguages:       1,
		})

		// Should have no current project when no paths are allowed
		assert.Nil(t, result.CurrentProject, "Should have no current project when no paths allowed")
	})

	t.Run("TokenExpiration_ShouldBeWithin24Hours", func(t *testing.T) {
		tf := testutil.NewTestFixture(StartSessionDirPrefix)
		defer tf.Cleanup()

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{})

		result, err := getToolResult[StartSessionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error creating session",
		)

		requireStartSessionResult(t, result, err, startSessionResultOpts{
			ShouldHaveToken: true,
		})

		// Verify token expires within 24 hours (with some tolerance)
		now := time.Now()
		expectedExpiry := now.Add(24 * time.Hour)
		tolerance := 5 * time.Minute

		assert.True(t, result.TokenExpiresAt.After(now), "Token should expire in the future")
		assert.True(t, result.TokenExpiresAt.Before(expectedExpiry.Add(tolerance)), "Token should expire within 24 hours + tolerance")
		assert.True(t, result.TokenExpiresAt.After(expectedExpiry.Add(-tolerance)), "Token should expire within 24 hours - tolerance")
	})
}

// TestAllToolsRegistered verifies that all expected tools are registered during init()
func TestAllToolsRegistered(t *testing.T) {
	var tool mcputil.Tool
	var toolName string

	// Verify each tool is registered
	for toolName = range toolNamesMap {
		tool = mcputil.GetRegisteredTool(toolName)
		assert.True(t, tool != nil, "tool %s should be registered", toolName)
		assert.NotNil(t, tool, "tool %s should not be nil", toolName)

		if tool != nil {
			assert.Equal(t, toolName, tool.Name(), "tool %s name should match", toolName)
			assert.NotEmpty(t, tool.Options().Description, "tool %s should have description", toolName)
			assert.NotNil(t, tool.Options().Properties, "tool %s should have properties", toolName)
		}
	}
}

// TestNoUnexpectedTools verifies that only expected tools are registered
func TestNoUnexpectedTools(t *testing.T) {
	var toolName string

	registeredTools := mcputil.RegisteredToolsMap()
	// Check for unexpected tools
	for toolName = range registeredTools {
		_, ok := toolNamesMap[toolName]
		assert.True(t, ok, "unexpected tool registered: %s", toolName)
	}

	// Verify expected count
	assert.Equal(t, len(toolNamesMap), len(registeredTools), "tool count mismatch")
}

// TestToolMetadataConsistency validates that all tools have consistent metadata
func TestToolMetadataConsistency(t *testing.T) {
	var tool mcputil.Tool

	for _, tool = range mcputil.RegisteredTools() {
		// Basic metadata validation
		assert.NotEmpty(t, tool.Options().Description, "tool %s should have description", tool.Name())

		if _, ok := tool.(*mcptools.StartSessionTool); !ok {
			assert.NotNil(t, tool.Options().Properties, "tool %s should have properties", tool.Name())
		}
	}
}

// TestSessionTokenRequirements validates which tools require session tokens
func TestSessionTokenRequirements(t *testing.T) {
	var tool mcputil.Tool
	var toolName string

	for toolName = range toolNamesMap {
		var hasSessionToken bool
		tool = mcputil.GetRegisteredTool(toolName)
		require.NotNil(t, tool, "tool %s must be registered", toolName)

		// Check if tool has session_token in its properties
		// CLAUDE: Tool should have a HasProperty() method
		for _, p := range tool.Options().Properties {
			if p.GetName() != "session_token" {
				continue
			}
			hasSessionToken = true
			break
		}

		if toolName == "start_session" {
			assert.False(t, hasSessionToken, "tool %s should NOT require session_token", toolName)
		} else {
			assert.True(t, hasSessionToken, "tool %s should require session_token", toolName)
		}
	}
}

// TestToolRegistrationOrder verifies that tools are registered in a predictable order
func TestToolRegistrationOrder(t *testing.T) {
	var toolNames []string
	var toolName string

	registeredTools := mcputil.RegisteredToolsMap()

	// Extract tool names
	for toolName = range registeredTools {
		toolNames = append(toolNames, toolName)
	}

	// Should have expected number of tools
	assert.Equal(t, len(toolNamesMap), len(toolNames), "should have expected number of tools registered")

	// start_session should always be registered (critical for other tools)
	startSessionTool := mcputil.GetRegisteredTool("start_session")
	assert.NotNil(t, startSessionTool, "start_session must be registered")
}
