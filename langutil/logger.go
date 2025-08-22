package langutil

import (
	"log/slog"
)

var logger *slog.Logger

func SetLogger(l *slog.Logger) {
	logger = l
}

func ensureLogger() {
	if logger == nil {
		panic("Must set logger with langutil.SetLogger() before using langutil package")
	}
}
