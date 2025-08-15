package test

import "testing"

// TestGetConfigToolWithJSONRPC tests the get_config tool via JSON-RPC.
func TestGetConfigToolWithJSONRPC(t *testing.T) {
	RunJSONRPCTest(t, nil, test{
		name: "get_config",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		subtests: map[string][]subtest{
			GoFile: {
				{
					arguments: sessionTokenArgs{},
					expected:  nil,
				},
			},
		},
	})
}
