package mcptools

import (
	"log/slog"
)

// logger is the package-level logger instance for mcptools.
var logger *slog.Logger

// SetLogger configures the package-level logger for mcptools.
func SetLogger(l *slog.Logger) {
	logger = l
}
