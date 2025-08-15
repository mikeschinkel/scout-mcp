package test

import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

// TestReplacePatternToolWithJSONRPC tests the replace_pattern tool via JSON-RPC.
func TestReplacePatternToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("replace-pattern-test")
	fixture.AddFileFixture("main.go", testutil.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "replace_pattern",
		arguments: replacePatternArgs{
			Path:           "main.go",
			Pattern:        "println",
			Replacement:    "fmt.Println",
			AllOccurrences: true,
		},
		expected: map[string]any{
			"jsonrpc":                               "2.0",
			"result.content.#":                      1,
			"result.content.0.type":                 "text",
			"result.content.0.text|json()|success":       true,
		},
	})
}
