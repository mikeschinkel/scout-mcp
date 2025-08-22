package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// searchFilesArgs represents arguments for the search_files tool.
type searchFilesArgs struct {
	Path        string   `json:"path"`
	Recursive   bool     `json:"recursive,omitempty"`
	Extensions  []string `json:"extensions,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	NamePattern string   `json:"name_pattern,omitempty"`
	FilesOnly   bool     `json:"files_only,omitempty"`
	DirsOnly    bool     `json:"dirs_only,omitempty"`
	MaxResults  int      `json:"max_results,omitempty"`
}

// TestSearchFilesToolWithJSONRPC tests the search_files tool via JSON-RPC.
func TestSearchFilesToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("search-files-test")
	fixture.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.AddFileFixture("index.js", &fsfix.FileFixtureArgs{
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
