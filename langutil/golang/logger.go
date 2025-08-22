package golang

import (
	"log/slog"

	"github.com/mikeschinkel/scout-mcp/langutil"
)

var logger *slog.Logger

func SetLogger(l *slog.Logger) {
	logger = l
	ensureLogger()
}

func ensureLogger() {
	if logger == nil {
		panic("Must set logger with golang.SetLogger() before using golang package")
	}
}

func init() {
	langutil.RegisterInitializerFunc(func(args langutil.Args) error {
		SetLogger(args.Logger)
		return nil
	})
}
