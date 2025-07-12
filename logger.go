package scout

import (
	"log/slog"
)

var logger = &slog.Logger{}

func GetLogger() *slog.Logger {
	return logger
}
func SetLogger(l *slog.Logger) *slog.Logger {
	logger = l
	return logger
}
