package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// analyzeFilesArgs represents arguments for the analyze_files tool.
type analyzeFilesArgs struct {
	Files []string `json:"files"`
}

// TestAnalyzeFilesToolWithJSONRPC tests the analyze_files tool via JSON-RPC.
func TestAnalyzeFilesToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("analyze-files-test")
	fixture.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
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
