package cliutil

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// OutputWriter defines the interface for user-facing output
type OutputWriter interface {
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Console writes to stdout/stderr for normal CLI usage
type outputWriter struct {
	stdout io.Writer
	stderr io.Writer
}

// NewOutputWriter creates a console output writer
func NewOutputWriter() OutputWriter {
	return &outputWriter{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Printf writes formatted output to stdout
func (c *outputWriter) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(c.stdout, format, args...)
}

// Errorf writes formatted error output to stderr
func (c *outputWriter) Errorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(c.stderr, format, args...)
}

// Package-level output instance
var output OutputWriter = NewOutputWriter()
var printMu sync.RWMutex
var errorMu sync.RWMutex

// SetOutput sets the global output writer (primarily for testing)
func SetOutput(writer OutputWriter) OutputWriter {
	printMu.Lock()
	defer printMu.Unlock()
	output = writer
	return output
}

// GetOutput returns the current output writer
func GetOutput() OutputWriter {
	printMu.RLock()
	defer printMu.RUnlock()
	return output
}

// Package-level convenience functions

// Printf writes formatted output
func Printf(format string, args ...interface{}) {
	printMu.RLock()
	defer printMu.RUnlock()
	output.Printf(format, args...)
}

// Errorf writes formatted error output
func Errorf(format string, args ...interface{}) {
	errorMu.RLock()
	defer errorMu.RUnlock()
	output.Errorf(format, args...)
}
