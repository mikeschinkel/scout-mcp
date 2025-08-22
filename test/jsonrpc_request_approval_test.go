package test

import "testing"

// requestApprovalArgs represents arguments for the request_approval tool.
type requestApprovalArgs struct {
	Operation      string   `json:"operation"`
	Files          []string `json:"files"`
	ImpactSummary  string   `json:"impact_summary,omitempty"`
	PreviewContent string   `json:"preview_content,omitempty"`
	RiskLevel      string   `json:"risk_level,omitempty"`
}

// TestRequestApprovalToolWithJSONRPC tests the request_approval tool via JSON-RPC.
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
