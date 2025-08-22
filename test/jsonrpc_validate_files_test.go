package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// validateFilesArgs represents arguments for the validate_files tool.
type validateFilesArgs struct {
	Paths      []string `json:"paths,omitempty"`
	Files      []string `json:"files,omitempty"`
	Language   string   `json:"language,omitempty"`
	Extensions []string `json:"extensions,omitempty"`
	Recursive  bool     `json:"recursive,omitempty"`
}

// TestValidateFilesToolWithJSONRPC tests the validate_files tool via JSON-RPC.
func TestValidateFilesToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("validate-files-test")
	fixture.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "validate_files",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		subtests: map[string][]subtest{
			GoFile: {
				{
					arguments: validateFilesArgs{
						Files: []string{"main.go"},
					},
					expected: nil,
				},
			},
		},
	})
}
