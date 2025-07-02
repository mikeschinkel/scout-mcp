package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func main() {
	var server *MCPServer
	var err error
	var input string
	var additionalPaths []string
	var opts MCPServerOpts
	var isInit bool
	var args ConfigArgs

	// Initialize logger
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	additionalPaths, opts, isInit, args, err = parseArgs()
	if err != nil {
		logger.Error("Error parsing arguments", "error", err)
		os.Exit(1)
	}

	if isInit {
		err = createDefaultConfig(args)
		if err != nil {
			logger.Error("Failed to create config", "error", err)
			os.Exit(1)
		}
		return
	}

	server, err = NewMCPServer(additionalPaths, opts)
	if err != nil {
		if len(additionalPaths) == 0 && !opts.OnlyMode {
			showUsageError()
			os.Exit(1)
		}
		logger.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	logger.Info("Scout-MCP File Search Server")
	logger.Info("Whitelisted directories:")
	for dir := range server.whitelistedDirs {
		logger.Info("Directory", "path", dir)
	}
	logger.Info("Server configuration", "port", server.config.Port)
	fmt.Printf("Press Enter to start server, or 'q' to quit: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "q" || input == "quit" {
		return
	}

	err = server.Start()
	if err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
