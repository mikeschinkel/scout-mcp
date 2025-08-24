package testutil

import (
	"bytes"
	"fmt"

	"github.com/mikeschinkel/scout-mcp/cliutil"
)

// TestOutputWriter captures output for testing
type TestOutputWriter struct {
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

// NewTestOutputWriter creates a test output writer that captures output in buffers
func NewTestOutputWriter() *TestOutputWriter {
	return &TestOutputWriter{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
}

// Printf writes formatted output to stdout buffer
func (t *TestOutputWriter) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(t.stdout, format, args...)
}

// Errorf writes formatted error output to stderr buffer
func (t *TestOutputWriter) Errorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(t.stderr, format, args...)
}

// StdoutString returns the captured stdout content as string
func (t *TestOutputWriter) StdoutString() string {
	return t.stdout.String()
}

// StderrString returns the captured stderr content as string
func (t *TestOutputWriter) StderrString() string {
	return t.stderr.String()
}

// Reset clears both buffers
func (t *TestOutputWriter) Reset() {
	t.stdout.Reset()
	t.stderr.Reset()
}

// Ensure TestOutputWriter implements OutputWriter interface
var _ cliutil.OutputWriter = (*TestOutputWriter)(nil)
