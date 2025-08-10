package scout

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mikeschinkel/scout-mcp/langutil"
	"github.com/mikeschinkel/scout-mcp/mcptools"
)

type RunArgs struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
}

func RunMain(ra RunArgs) (err error) {
	var server *MCPServer
	var args Args
	var serverOpts Opts

	// Initialize logger to file
	err = InitializeFileLogger()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		goto end
	}
	logger.Info("CLI Args:", "args", os.Args[1:])

	// Initialize logger to file
	err = Initialize()
	if err != nil {
		logger.Error("Failed to initialize %s: %v", AppName, err)
		goto end
	}

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

	err = server.StartMCP()
	if err != nil {
		logger.Error("MCP server failed", "error", err)
		goto end
	}

end:
	return err
}

func Initialize() (err error) {
	return langutil.Initialize(langutil.Args{
		AppName: AppName,
	})
}

func InitializeFileLogger() (err error) {
	var logger *slog.Logger
	var logDir string
	var logPath string
	var logFile *os.File
	var homeDir string

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	logDir = filepath.Join(homeDir, "Library", "Logs", "scout-mcp")
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		goto end
	}

	logPath = filepath.Join(logDir, "scout-mcp.log")
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		goto end
	}

	logger = slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Use JSON logging for structured logs
	SetLogger(logger)
	mcptools.SetLogger(logger)

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
