package scout

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func RunMain() (err error) {
	var server *MCPServer
	var args Args
	var opts Opts

	// Initialize logger to file
	err = InitializeFileLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		goto end
	}

	args, err = ParseArgs()
	if err != nil {
		logger.Error("Error parsing arguments", "error", err)
		goto end
	}

	if args.IsInit {
		err = CreateDefaultConfig(args.ConfigArgs)
		if err != nil {
			logger.Error("Failed to create config", "error", err)
			goto end
		}
		goto end
	}

	// Convert MCPServerOpts to Opts
	opts = Opts{
		OnlyMode: args.ServerOpts.OnlyMode,
	}

	server, err = NewMCPServer(args.AdditionalPaths, opts)
	if err != nil {
		if len(args.AdditionalPaths) == 0 && !opts.OnlyMode {
			ShowUsageError()
			goto end
		}
		logger.Error("Failed to create server", "error", err)
		goto end
	}

	logger.Info("Scout-MCP File Operations Server starting")
	logger.Info("Whitelisted directories:")
	for dir := range server.WhitelistedDirs() {
		logger.Info("Directory", "path", dir)
	}

	err = server.StartMCP()
	if err != nil {
		logger.Error("MCP server failed", "error", err)
		goto end
	}

end:
	return err
}

func InitializeFileLogger() (err error) {
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

	// Use JSON logging for structured logs
	SetLogger(slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	logger.Info("Logger initialized", "log_file", logPath)

end:
	return err
}

func ShowUsageError() {
	var displayPath string

	displayPath = filepath.Join("~", ConfigBaseDirName, ConfigDirName, ConfigFileName)

	fmt.Printf(`ERROR: No whitelisted directories specified.

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
`, AppName, displayPath)
}
