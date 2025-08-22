package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// updateFileArgs represents arguments for the update_file tool.
type updateFileArgs struct {
	Filepath   string `json:"filepath"`
	NewContent string `json:"new_content"`
}

// TestUpdateFileToolWithJSONRPC tests the update_file tool via JSON-RPC.
func TestUpdateFileToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("update-file-test")
	fixture.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.AddFileFixture("index.js", &fsfix.FileFixtureArgs{
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
