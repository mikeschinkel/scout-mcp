package test

// getJSONRPCRequestApprovalTest returns the test definition for request_approval tool
import "testing"

func TestRequestApprovalToolWithJSONRPC(t *testing.T) {
	RunJSONRPCTest(t, nil, test{
		name: "request_approval",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		subtests: map[string][]subtest{
			GoFile: {
				{
					arguments: requestApprovalArgs{
						Operation:     "test operation",
						Files:         []string{"test.txt"},
						ImpactSummary: "Test approval request",
						RiskLevel:     "low",
					},
					expected: nil,
				},
			},
		},
	})
}
