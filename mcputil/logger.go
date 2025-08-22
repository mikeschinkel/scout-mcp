package mcputil

import (
	"log/slog"
)

// logger is the package-level logger instance for mcputil.
var logger *slog.Logger

// SetLogger configures the package-level logger for mcputil.
// This function allows the application to provide a structured logger
// for mcputil operations including tool execution, session management,
// and error reporting throughout the MCP server framework.
func SetLogger(l *slog.Logger) {
	logger = l
}
