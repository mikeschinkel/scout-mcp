package test

import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

// TestSearchFilesToolWithJSONRPC tests the search_files tool via JSON-RPC.
func TestSearchFilesToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("search-files-test")
	fixture.AddFileFixture("main.go", testutil.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.AddFileFixture("index.js", testutil.FileFixtureArgs{
		Content: "console.log('Hello');\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "search_files",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		subtests: map[string][]subtest{
			GoFile: {
				{
					arguments: searchFilesArgs{
						Path: ".",
					},
					expected: nil,
				},
			},
		},
	})
}
