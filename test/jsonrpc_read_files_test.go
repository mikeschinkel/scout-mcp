package test

import (
	"testing"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

func TestReadFilesToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("read-files-test")
	fixture.AddFileFixture("main.go", testutil.FileFixtureArgs{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
	})
	fixture.AddFileFixture("index.js", testutil.FileFixtureArgs{
		Content: "console.log('Hello');\n",
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "read_files",
		arguments: readFilesArgs{
			Paths: []string{"main.go", "index.js"},
		},
		expected: map[string]any{
			"jsonrpc":                             "2.0",
			"result.content.#":                    1,
			"result.content.0.type":               "text",
			"result.content.0.text|json()|total_files": 2,
			"result.content.0.text|json()|files.#":     2,
			"result.content.0.text|json()|files.0.content": "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
			"result.content.0.text|json()|files.1.content": "console.log('Hello');\n",
			"result.content.0.text|json()|files.0.name":    "main.go",
			"result.content.0.text|json()|files.1.name":    "index.js",
		},
	})
}
