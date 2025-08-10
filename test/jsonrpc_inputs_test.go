package test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/mikeschinkel/scout-mcp"
)

// TestStartSession tests starting a session using JSONRPC input
func TestStartSessionWithJSONRPC(t *testing.T) {
	// The exact JSON that causes the panic
	input := `{"jsonrpc": "2.0","id": 1,"method": "tools/call","params": {"name": "start_session","arguments": {}}}` + "\n"

	// Create readers/writers for stdio
	stdin := strings.NewReader(input)
	stdout := &bytes.Buffer{}

	// Set breakpoint on the next line in GoLand and step through
	// This will call the exact same code path that panics from CLI
	err := scout.RunMain(scout.RunArgs{
		Args:   os.Args[:1],
		Stdin:  stdin,
		Stdout: stdout,
	})
	if err != nil {
		t.Error(err.Error())
	}
	// TODO Add a test that the results is what is expected.
	//t.Logf("STDOUT: %s", stdout.String())
}
