package test

import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

// TestUpdateFileToolWithJSONRPC tests the update_file tool via JSON-RPC.
func TestUpdateFileToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("update-file-test")
	fixture.AddFileFixture("main.go", testutil.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.AddFileFixture("index.js", testutil.FileFixtureArgs{
		Content: "console.log('Hello');\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "update_file",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		subtests: map[string][]subtest{
			GoFile: {
				{
					arguments: updateFileArgs{
						Filepath:   "main.go",
						NewContent: "package main\n\nfunc main() {\n\tprintln(\"Updated!\")\n}",
					},
					expected: nil,
				},
			},
			JavascriptFile: {
				{
					arguments: updateFileArgs{
						Filepath:   "index.js",
						NewContent: "console.log('Updated!');",
					},
					expected: nil,
				},
			},
		},
	})
}
