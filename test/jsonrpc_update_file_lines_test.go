package test

import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

// TestUpdateFileLinesToolWithJSONRPC tests the update_file_lines tool via JSON-RPC.
func TestUpdateFileLinesToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("update-file-lines-test")
	fixture.AddFileFixture("main.go", testutil.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "update_file_lines",
		arguments: updateFileLinesArgs{
			Filepath:   "main.go",
			NewContent: "\tfmt.Println(\"Updated!\")",
			StartLine:  4,
			EndLine:    4,
		},
		expected: map[string]any{
			"jsonrpc":                               "2.0",
			"result.content.#":                      1,
			"result.content.0.type":                 "text",
			"result.content.0.text|json()|success":       true,
			"result.content.0.text|json()|start_line":    4,
			"result.content.0.text|json()|end_line":      4,
		},
	})
}
