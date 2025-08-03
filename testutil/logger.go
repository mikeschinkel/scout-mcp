package testutil

import (
	"context"
	"log/slog"
)

func NewTestLogger() *slog.Logger {
	return QuietLogger() // TODO: Replace this with a buffering logger
}

// QuietLogger creates a logger that discards all output (for tests that don't need log inspection)
func QuietLogger() *slog.Logger {
	return slog.New(NullHandler{})
}

type NullHandler struct{}

func (NullHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (NullHandler) Handle(context.Context, slog.Record) error { return nil }
func (NullHandler) WithAttrs([]slog.Attr) slog.Handler        { return NullHandler{} }
func (NullHandler) WithGroup(string) slog.Handler             { return NullHandler{} }

//// QuietLogger creates a logger that discards all output (for tests that don't need log inspection)
//func QuietLogger() *slog.Logger {
//	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
//		Level: slog.LevelError + 1, // Set level higher than any used level to discard everything
//	}))
//}
