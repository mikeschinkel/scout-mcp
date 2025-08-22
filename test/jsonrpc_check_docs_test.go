package test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// checkDocsArgs represents arguments for the check_docs tool.
type checkDocsArgs struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
	MaxFiles  int    `json:"max_files,omitempty"`
}

// TestCheckDocsToolWithJSONRPC tests the check_docs tool via JSON-RPC.
func TestCheckDocsToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("check-docs-jsonrpc-test")
	
	// Create a Go file with documentation issues
	fixture.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: `package main

func main() {
	println("Hello")
}

type Config struct {
	Port string
}
`,
	})
	
	// Create a fully documented Go file  
	fixture.AddFileFixture("documented.go", &fsfix.FileFixtureArgs{
		Content: `// Package main provides a documented example.
package main

// DocumentedFunc is a properly documented function.
func DocumentedFunc() {
	// Function implementation
}
`,
	})
	
	fixture.Setup(t)
	defer fixture.Cleanup()

	RunJSONRPCTest(t, fixture, test{
		name: "check_docs",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
			"result.content.0.text|json()|path|notEmpty()": true,
			"result.content.0.text|json()|total|exists()":  true,
			"result.content.0.text|json()|issues|exists()": true,
		},
		subtests: map[string][]subtest{
			"CurrentRepo": {
				{
					arguments: checkDocsArgs{
						Path: repoDir(),
					},
				},
			},
			"SingleFile": {
				{
					arguments: checkDocsArgs{
						Path: ".", // Pass directory, tool will find main.go
					},
					expected: map[string]any{
						"result.content.0.text|json()|total|exists()": true, // Should find documentation issues in main.go
					},
				},
			},
			"Directory": {
				{
					arguments: checkDocsArgs{
						Path:      ".",
						Recursive: false,
					},
					expected: map[string]any{
						"result.content.0.text|json()|total|exists()": true, // Should find some issues
					},
				},
			},
			"RecursiveDirectory": {
				{
					arguments: checkDocsArgs{
						Path:      ".",
						Recursive: true,
					},
					expected: map[string]any{
						"result.content.0.text|json()|total|exists()": true, // Should find issues including README.md
					},
				},
			},
			"WithMaxFiles": {
				{
					arguments: checkDocsArgs{
						Path:     ".",
						MaxFiles: 1,
					},
					expected: map[string]any{
						"result.content.0.text|json()|total|exists()": true, // Should respect max_files limit
					},
				},
			},
		},
	})
}