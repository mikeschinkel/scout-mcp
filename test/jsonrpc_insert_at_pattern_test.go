package test

import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

// TestInsertAtPatternToolWithJSONRPC tests the insert_at_pattern tool via JSON-RPC.
func TestInsertAtPatternToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("insert-at-pattern-test")
	fixture.AddFileFixture("main.go", testutil.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "insert_at_pattern",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		subtests: map[string][]subtest{
			GoFile: {
				{
					arguments: insertAtPatternArgs{
						Path:          "main.go",
						NewContent:    "// Added before main",
						BeforePattern: "func main",
						Position:      "before",
					},
					expected: nil,
				},
			},
		},
	})
}
