package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// findFilePartArgs represents arguments for the find_file_part tool.
type findFilePartArgs struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	PartType string `json:"part_type"`
	PartName string `json:"part_name"`
}

// TestFindFilePartToolWithJSONRPC tests the find_file_part tool via JSON-RPC.
func TestFindFilePartToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("find-file-part-test")
	fixture.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "find_file_part",
		arguments: findFilePartArgs{
			Path:     "main.go",
			Language: "go",
			PartType: "func",
			PartName: "main",
		},
		expected: map[string]any{
			"jsonrpc":                               "2.0",
			"result.content.#":                      1,
			"result.content.0.type":                 "text",
			"result.content.0.text|json()|found":         true,
			"result.content.0.text|json()|part_name":     "main",
			"result.content.0.text|json()|part_type":     "func",
		},
	})
}
