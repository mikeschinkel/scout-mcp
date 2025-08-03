package mcputil

import (
	"log/slog"
)

var logger *slog.Logger

func SetLogger(l *slog.Logger) {
	logger = l
}
