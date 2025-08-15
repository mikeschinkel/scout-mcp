package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// TestJSONRPCProtocolFailures tests that malformed JSON-RPC requests are handled properly
func TestJSONRPCProtocolFailures(t *testing.T) {
	// Test invalid tool name - should return JSON-RPC error response
	t.Run("invalid_tool_name", func(t *testing.T) {
		RunJSONRPCTest(t, nil, test{
			name: "nonexistent_tool",
			cliArgs: []string{"/tmp"}, // Use /tmp as basic allowed path
			expected: map[string]any{
				"jsonrpc":                  "2.0",
				"error.code":               mcputil.JSONRPCInvalidParams, // -32602: Invalid params per JSON-RPC 2.0 spec
				"error.message|notEmpty()": true,                         // Should have an error message
			},
		})
	})
	
	// Test missing session token for tools that require it
	t.Run("missing_session_token", func(t *testing.T) {
		RunJSONRPCTest(t, nil, test{
			name: "read_files",
			arguments: map[string]any{}, // Deliberately omit session_token
			cliArgs: []string{"/tmp"},
			wantErr: true,
			expected: map[string]any{
				"jsonrpc": "2.0", 
				"result.isError": true,
				"result.content.0.text|notEmpty()": true, // Should have error message
			},
		})
	})
	
	// Test invalid arguments for valid tool
	t.Run("invalid_arguments", func(t *testing.T) {
		RunJSONRPCTest(t, nil, test{
			name: "read_files",
			arguments: map[string]any{
				"session_token": "valid_token",
				"invalid_param": "should_not_exist",
			},
			cliArgs: []string{"/tmp"},
			wantErr: true,
			expected: map[string]any{
				"jsonrpc": "2.0",
				"result.isError": true,
			},
		})
	})
}