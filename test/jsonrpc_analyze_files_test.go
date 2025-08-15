package test

import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

// getJSONRPCAnalyzeFilesTest returns the test definition for analyze_files tool
func TestAnalyzeFilesToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("analyze-files-test")
	fixture.AddFileFixture("main.go", testutil.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "analyze_files",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		arguments: analyzeFilesArgs{
			Files: []string{"main.go"},
		},
	})
}
