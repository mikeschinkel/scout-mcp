package scout

import (
	"log/slog"
)

// logger is the package-level logger instance used throughout the scout package.
var logger = &slog.Logger{}

// GetLogger returns the current package-level logger instance.
func GetLogger() *slog.Logger {
	return logger
}

// SetLogger sets the package-level logger and returns the new logger instance.
func SetLogger(l *slog.Logger) *slog.Logger {
	logger = l
	return logger
}

// ensureLogger validates that a logger has been configured and panics with
// a descriptive message if no logger is available. This function is called
// internally by package functions that require logging capabilities.
//
// This fail-fast approach ensures that missing logger configuration is
// detected early in the application lifecycle rather than causing silent
// failures or unexpected behavior during runtime operations.
//
// The panic message provides clear guidance on how to resolve the issue
// by calling SetLogger with a properly configured logger instance.
//
// Panics if logger is nil, indicating that SetLogger has not been called
// or was called with a nil logger instance.
//
// This function is used internally by:
//   - Save operations that need to log JSON marshaling or file write errors
//   - File close operations in mustClose that need to log close errors
//   - Any other operations that may need to report errors or debug information
func ensureLogger() {
	if logger == nil {
		panic("Must set logger with scout.SetLogger() before using scoutcfg package")
	}
}
