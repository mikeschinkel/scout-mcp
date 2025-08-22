package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// replacePatternArgs represents arguments for the replace_pattern tool.
type replacePatternArgs struct {
	Path           string `json:"path"`
	Pattern        string `json:"pattern"`
	Replacement    string `json:"replacement"`
	Regex          bool   `json:"regex,omitempty"`
	AllOccurrences bool   `json:"all_occurrences,omitempty"`
}

// TestReplacePatternToolWithJSONRPC tests the replace_pattern tool via JSON-RPC.
func TestReplacePatternToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("replace-pattern-test")
	fixture.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
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
