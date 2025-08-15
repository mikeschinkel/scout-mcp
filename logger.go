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
