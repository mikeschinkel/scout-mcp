package test

// getJSONRPCStartSessionTest returns the test definition for start_session tool
import "testing"

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
