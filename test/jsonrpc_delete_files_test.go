package test

import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

// getJSONRPCDeleteFilesTest returns the test definition for delete_files tool
func TestDeleteFilesToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("delete-files-test")
	fixture.AddFileFixture("test.go", testutil.FileFixtureArgs{
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
