package scout

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mikeschinkel/scout-mcp/langutil"
	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

type RunArgs struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
}

func RunMain(ctx context.Context, ra RunArgs) (err error) {
	var server *MCPServer
	var args Args
	var serverOpts Opts

	// Initialize scout package
	err = Initialize()
	if err != nil {
		logger.Error("Failed to initialize %s: %v", AppName, err)
		goto end
	}

	logger.Info("CLI Args:", "args", os.Args[1:])

	args, err = ParseArgs(ra.Args[1:])
	if err != nil {
		logger.Error("Error parsing arguments", "error", err)
		goto end
	}

	if args.IsInit {
		err = CreateDefaultConfig(args)
		if err != nil {
			logger.Error("Failed to create config", "error", err)
			goto end
		}
		goto end
	}

	// Convert MCPServerOpts to Opts
	serverOpts = Opts{
		OnlyMode:        args.ServerOpts.OnlyMode,
		AdditionalPaths: args.AdditionalPaths,
		Stdin:           ra.Stdin,
		Stdout:          ra.Stdout,
	}

	server, err = NewMCPServer(serverOpts)
	if err != nil {
		if len(serverOpts.AdditionalPaths) == 0 && !serverOpts.OnlyMode {
			ShowUsageError(err)
			goto end
		}

		logger.Error("Failed to create server", "error", err)
		goto end
	}

	logger.Info("Scout-MCP File Operations Server starting")
	logger.Info("Allowed directories:")
	for path := range server.AllowedPaths() {
		logger.Info("Directory", "path", path)
	}

	err = server.StartMCP(ctx)
	if err != nil {
		logger.Error("MCP server failed", "error", err)
		goto end
	}

end:
	return err
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
		Logger: logger,
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
