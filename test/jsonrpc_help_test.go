package test

import "testing"

// TestHelpToolWithJSONRPC tests the help tool via JSON-RPC.
func TestHelpToolWithJSONRPC(t *testing.T) {
	RunJSONRPCTest(t, nil, test{
		name: "help",
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
