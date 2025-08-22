package scoutcfg

import (
	"log/slog"
)

// logger is the package-level logger instance used by all FileStore operations.
// It must be initialized with SetLogger before any FileStore operations that
// require logging (such as Save operations or error handling in mustClose).
//
// The logger is used for:
//   - Error logging during file close operations
//   - Debug information during configuration operations
//   - Tracking configuration file access patterns
//
// This global approach ensures consistent logging behavior across all
// FileStore instances while allowing the application to control the
// logging configuration and destination.
var logger *slog.Logger

// SetLogger configures the slog.Logger instance that will be used by the
// scoutcfg package for all logging operations. This function must be called
// before performing any FileStore operations that require logging.
//
// The logger is used throughout the package for error reporting, debug
// information, and operational tracking. Setting a logger is mandatory
// for proper package functionality - operations that require logging
// will panic if no logger has been configured.
//
// Parameters:
//   - l: A configured slog.Logger instance. This logger will be used for
//     all subsequent logging operations in the package. The logger should
//     be properly configured with appropriate log levels and output
//     destinations before being passed to this function.
//
// Example usage:
//
//	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
//		Level: slog.LevelInfo,
//	}))
//	scoutcfg.SetLogger(logger)
//
// This function should typically be called during application initialization,
// before creating any FileStore instances or performing configuration
// operations.
func SetLogger(l *slog.Logger) {
	logger = l
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
		panic("Must set logger with scoutcfg.SetLogger() before using scoutcfg package")
	}
}