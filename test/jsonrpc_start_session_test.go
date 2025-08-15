package test

import "testing"

// TestStartSessionToolWithJSONRPC tests the start_session tool via JSON-RPC.
func TestStartSessionToolWithJSONRPC(t *testing.T) {
	RunJSONRPCTest(t, nil, test{
		name: "start_session",
		expected: map[string]any{
			"jsonrpc":                         "2.0",
			"result.content.#":                1,
			"result.content.0.type":           "text",
			"result.content.0.text|json()|message": "MCP Session Started",
		},
	})
}
