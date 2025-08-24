package scoutcmds

import (
	"log/slog"
	"os"
)

// SetupLogger initializes the logger based on verbosity
func SetupLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}
