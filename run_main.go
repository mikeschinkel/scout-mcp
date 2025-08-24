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
	MCPInput       io.Reader
	MCPOutput      io.Writer
	CLIOutput      cliutil.OutputWriter
	ConfigProvider ConfigProvider
}

func RunMain(ctx context.Context, ra RunArgs) (err error) {
	var runner *cliutil.CmdRunner
	var ctxWithCancel context.Context
	var cancel context.CancelFunc

	// Initialize Scout
	err = Initialize()
	if err != nil {
		logger.Error("Failed to initialize Scout", "error", err)
		goto end
	}

	// Initialize CLI framework
	err = cliutil.Initialize()
	if err != nil {
		logger.Error("Failed to initialize CLI framework", "error", err)
		goto end
	}

	// Set up signal handling for the context
	ctxWithCancel, cancel = setupSignalHandling(ctx, logger)
	defer cancel()

	ra.ConfigProvider.SetIO(ra.MCPInput, ra.MCPOutput)

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
		ra.CLIOutput.Errorf("Error: %v\n", err)
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

func Initialize() (err error) {

	err = initializeFileLogger(logger)
	if err != nil {
		err = fmt.Errorf("failed to initialize logger: %v\n", err)
		goto end
	}

	logger.Info("Logger initialized\n")

	err = langutil.Initialize(langutil.Args{
		AppName: AppName,
		Logger:  logger,
	})

	logger.Info("langutil initialized\n")

end:
	return err
}

func initializeFileLogger(logger *slog.Logger) (err error) {
	var logDir string
	var logPath string
	var logFile *os.File
	var homeDir string

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	logDir = filepath.Join(homeDir, ".config", "scout-mcp")
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		goto end
	}
	logPath = filepath.Join(logDir, "errors.log")
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		goto end
	}

	logger = slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Use JSON logging for structured logs
	SetLogger(logger)

	//TODO These should be registered by the package, not hard-coded here.
	mcptools.SetLogger(logger)
	mcputil.SetLogger(logger)

	logger.Info("Logger initialized", "log_file", logPath)

end:
	return err
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
