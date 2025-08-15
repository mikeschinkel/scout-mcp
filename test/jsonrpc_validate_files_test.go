package test

// getJSONRPCValidateFilesTest returns the test definition for validate_files tool
import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

func TestValidateFilesToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("validate-files-test")
	fixture.AddFileFixture("main.go", testutil.FileFixtureArgs{
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
