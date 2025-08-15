package test

import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

func TestCreateFileToolWithJSONRPC(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		content  string
	}{
		{
			name:     "create Go file",
			filepath: "main.go",
			content:  "package main\n\nimport (\n\t\"log\"\n)\n\nfunc main() {\n\tlog.Println(\"Claude Code LLM Gateway Proxy starting...\")\n}",
		},
		{
			name:     "create JavaScript file",
			filepath: "index.js",
			content:  "console.log('Hello, World!');",
		},
		{
			name:     "create Python file",
			filepath: "main.py",
			content:  "#!/usr/bin/env python3\nprint('Hello, World!')",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := testutil.NewTestFixture("create-file-test")
			fixture.Setup(t)
			defer fixture.Cleanup()
			
			RunJSONRPCTest(t, fixture, test{
				name: "create_file",
				arguments: filepathContent{
					Filepath:   tc.filepath,
					NewContent: tc.content,
				},
				expected: map[string]any{
					"jsonrpc":                         "2.0",
					"result.content.#":                1,
					"result.content.0.type":           "text",
					"result.content.0.text|json()|success": true,
				},
			})
		})
	}
}
