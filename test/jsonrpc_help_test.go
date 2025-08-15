package test

// getJSONRPCHelpTest returns the test definition for help tool
import "testing"

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
