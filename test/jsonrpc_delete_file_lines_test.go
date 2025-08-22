package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// deleteFileLinesArgs represents arguments for the delete_file_lines tool.
type deleteFileLinesArgs struct {
	Filepath  string `json:"filepath"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// TestDeleteFileLinesToolWithJSONRPC runs the test for the 'delete_file_lines' MCP Server tool.
func TestDeleteFileLinesToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("delete-file-lines-test")
	fixture.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: "package main\n\n// Comment to delete\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()

	RunJSONRPCTest(t, fixture, test{
		name: "delete_file_lines",
		arguments: deleteFileLinesArgs{
			Filepath:  "main.go",
			StartLine: 3,
			EndLine:   3,
		},
		expected: map[string]any{
			"jsonrpc":                                 "2.0",
			"result.content.#":                        1,
			"result.content.0.type":                   "text",
			"result.content.0.text|json()|success":    true,
			"result.content.0.text|json()|start_line": 3,
			"result.content.0.text|json()|end_line":   3,
		},
	})
}
