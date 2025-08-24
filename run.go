package scout

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mikeschinkel/scout-mcp/cliutil"
	"github.com/mikeschinkel/scout-mcp/langutil"
	"github.com/mikeschinkel/scout-mcp/langutil/golang"
	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// ConfigProvider provides access to CLI commands and configuration
type ConfigProvider interface {
	GetConfig() cliutil.Config
	GlobalFlagSet() *cliutil.FlagSet
	SetIO(io.Reader, io.Writer)
	GetIO() (io.Reader, io.Writer)
}

type RunArgs struct {
	Args           []string
	MCPReader      io.Reader
	MCPWriter      io.Writer
	CLIWriter      cliutil.OutputWriter
	ConfigProvider ConfigProvider
	Logger         *slog.Logger
}

func Run(ctx context.Context, ra RunArgs) (err error) {
	var runner *cliutil.CmdRunner
	var ctxWithCancel context.Context
	var cancel context.CancelFunc

	// Initialize Scout
	err = Initialize(ra.Logger)
	if err != nil {
		logger.Error("Failed to initialize Scout", "error", err)
		goto end
	}

	// Initialize CLI framework
	err = cliutil.Initialize(ra.CLIWriter)
	if err != nil {
		logger.Error("Failed to initialize CLI framework", "error", err)
		goto end
	}

	// Set up signal handling for the context
	ctxWithCancel, cancel = setupSignalHandling(ctx, logger)
	defer cancel()

	ra.ConfigProvider.SetIO(ra.MCPReader, ra.MCPWriter)

	// Set up command runner
	runner = cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
		Config:        ra.ConfigProvider.GetConfig(),
		GlobalFlagSet: ra.ConfigProvider.GlobalFlagSet(),
		Args:          ra.Args[1:], // Skip program name
	})

	// Execute command
	err = runner.Run(ctxWithCancel)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Info("Operation cancelled by user")
			goto end
		}
		cliutil.Errorf("Error: %v\n", err)
		logger.Error("Command failed", "error", err)
		goto end
	}

end:
	return err
}

func setupSignalHandling(ctx context.Context, logger *slog.Logger) (context.Context, context.CancelFunc) {
	ctxWithCancel, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received interrupt signal, shutting down...")
		cancel()
	}()

	return ctxWithCancel, cancel
}

func Initialize(logger *slog.Logger) (err error) {

	initializeLoggers(logger)

	err = langutil.Initialize(langutil.Args{
		AppName: AppName,
		Logger:  logger,
	})

	//end:
	return err
}

func CreateJSONLogger() (logger *slog.Logger, err error) {
	var logDir string
	var logFilePath string
	var homeDir string
	var logFile *os.File

	homeDir, err = os.UserHomeDir()
	if err != nil {
		err = fmt.Errorf("failed to access home directory; %w\n", err)
		goto end
	}

	logDir = filepath.Join(homeDir, ".config", "scout-mcp")
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		err = fmt.Errorf("failed to make log directory %s; %w\n", logDir, err)
		goto end
	}
	logFilePath = filepath.Join(logDir, "errors.log")
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		err = fmt.Errorf("failed to open log logFile %s; %w\n", logFilePath, err)
		goto end
	}
	defer mustClose(logFile)

	logger = slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Logger initialized", "log_file", logFilePath)

end:
	if err != nil {
		err = fmt.Errorf("failed to initialize logger; %w\n", err)
	}
	return logger, err
}

func initializeLoggers(logger *slog.Logger) {
	//TODO These should be registered by the package, not hard-coded here.
	// OR if possible resolved via reflection
	SetLogger(logger)
	mcptools.SetLogger(logger)
	mcputil.SetLogger(logger)
	cliutil.SetLogger(logger)
	langutil.SetLogger(logger)
	golang.SetLogger(logger)
}

func ShowUsageError(err error) {
	var displayPath string

	displayPath = filepath.Join("~", ConfigBaseDirName, ConfigDirName, ConfigFileName)

	fmt.Printf(`ERROR: %[3]s.

Usage options:
  %[1]s <path>              Add path to config file paths
  %[1]s --only <path>       Use only the specified path
  %[1]s init                Create empty config file
  %[1]s init <path>         Create config with custom initial path

Config file location: %[2]s

Examples:
  %[1]s ~/MyProjects
  %[1]s --only /tmp/safe-dir
  %[1]s init ~/Code
`, AppName, displayPath, err.Error())
}
