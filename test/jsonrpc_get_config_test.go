package test

// getJSONRPCGetConfigTest returns the test definition for get_config tool
import "testing"

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
