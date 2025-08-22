package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// deleteFilesArgs represents arguments for the delete_files tool.
type deleteFilesArgs struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
}

// TestDeleteFilesToolWithJSONRPC tests the delete_files tool via JSON-RPC.
func TestDeleteFilesToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("delete-files-test")
	fixture.AddFileFixture("test.go", &fsfix.FileFixtureArgs{
		Content: "package test\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "delete_files",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		subtests: map[string][]subtest{
			GoFile: {
				{
					arguments: deleteFilesArgs{
						Path:      "test.go",
						Recursive: false,
					},
					expected: nil,
				},
			},
		},
	})
}
