package test

import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

// TestReplaceFilePartToolWithJSONRPC tests the replace_file_part tool via JSON-RPC.
func TestReplaceFilePartToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("replace-file-part-test")
	fixture.AddFileFixture("main.go", testutil.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "replace_file_part",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		subtests: map[string][]subtest{
			GoFile: {
				{
					arguments: replaceFilePartArgs{
						Path:       "main.go",
						Language:   "go",
						PartType:   "func",
						PartName:   "main",
						NewContent: "func main() {\n\tfmt.Println(\"Replaced!\")\n}",
					},
					expected: nil,
				},
			},
		},
	})
}
