package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// replaceFilePartArgs represents arguments for the replace_file_part tool.
type replaceFilePartArgs struct {
	Path       string `json:"path"`
	Language   string `json:"language"`
	PartType   string `json:"part_type"`
	PartName   string `json:"part_name"`
	NewContent string `json:"new_content"`
}

// TestReplaceFilePartToolWithJSONRPC tests the replace_file_part tool via JSON-RPC.
func TestReplaceFilePartToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("replace-file-part-test")
	fixture.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
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
