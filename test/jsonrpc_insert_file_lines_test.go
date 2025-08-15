package test

// getJSONRPCInsertFileLinesTest returns the test definition for insert_file_lines tool
import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

func TestInsertFileLinesToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("insert-file-lines-test")
	fixture.AddFileFixture("main.go", testutil.FileFixtureArgs{
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
