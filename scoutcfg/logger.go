package scoutcfg

import (
	"log/slog"
)

var logger *slog.Logger

// SetLogger sets the slog.Logger to use
func SetLogger(l *slog.Logger) {
	logger = l
}

// ensureLogger panics if logger is not set
func ensureLogger() {
	if logger == nil {
		panic("Must set logger with gmover.SetLogger() before using gmover package")
	}
}
