package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// insertFileLinesArgs represents arguments for the insert_file_lines tool.
type insertFileLinesArgs struct {
	Filepath   string `json:"filepath"`
	NewContent string `json:"new_content"`
	Position   string `json:"position"`
	LineNumber int    `json:"line_number"`
}

// TestInsertFileLinesToolWithJSONRPC tests the insert_file_lines tool via JSON-RPC.
func TestInsertFileLinesToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("insert-file-lines-test")
	fixture.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "insert_file_lines",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		subtests: map[string][]subtest{
			GoFile: {
				{
					arguments: insertFileLinesArgs{
						Filepath:   "main.go",
						NewContent: "// Inserted comment",
						Position:   "before",
						LineNumber: 1,
					},
					expected: nil,
				},
			},
		},
	})
}
