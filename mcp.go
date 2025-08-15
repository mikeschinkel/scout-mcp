// Package scout provides a secure Model Context Protocol (MCP) server
// for file operations with explicit directory whitelisting and session-based
// instruction enforcement. The server enables Claude to perform safe file
// operations through stdio transport with comprehensive approval workflows.
package scout

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// MCPServer represents a Model Context Protocol server instance that provides
// secure file operations to Claude through stdio transport. It manages
// directory whitelisting, session tokens, and user approval workflows.
type MCPServer struct {
	config          *Config
	additionalPaths map[string]struct{}
	allowedPaths    map[string]struct{}
	mcpServer       mcputil.Server
	Stdin           io.Reader
	StdOut          io.Writer
}

// NewMCPServer creates a new MCP server instance with the given options.
// It loads configuration, validates paths, and registers all available tools.
// The server is ready to start serving requests via stdio transport.
func NewMCPServer(opts Opts) (s *MCPServer, err error) {

	s = &MCPServer{
		additionalPaths: toExistenceMap(opts.AdditionalPaths),
	}

	// Load config and validate paths
	s.config, err = s.loadConfig(opts)
	if err != nil {
		goto end
	}

	// Create MCP server with stdio transport using mcputil
	s.mcpServer = mcputil.NewServer(mcputil.ServerOpts{
		Name:        AppName,
		Version:     AppVersion,
		Tools:       true,
		Subscribe:   false,
		ListChanged: false,
		Prompts:     false,
		Logging:     true,
		Stdin:       opts.Stdin,
		Stdout:      opts.Stdout,
	})

	// Register tools
	err = s.registerTools()
	if err != nil {
		goto end
	}

end:
	return s, err
}

// StartMCP starts the MCP server and begins serving requests via stdio transport.
// This is a blocking call that runs until the context is canceled or an error occurs.
func (s *MCPServer) StartMCP(ctx context.Context) (err error) {
	logger.Info("MCP server starting with stdio transport")
	logger.Info("Allowed directories:")
	for dir := range s.AllowedPaths() {
		logger.Info("Directory", "path", dir)
	}

	// Start the stdio server (blocking)
	err = s.mcpServer.ServeStdio(ctx)

	return err
}

// AllowedPaths returns a map of all directories that the server is allowed to access.
// This includes both paths from the configuration file and additional paths provided at startup.
func (s *MCPServer) AllowedPaths() map[string]struct{} {
	if s.allowedPaths != nil {
		goto end
	}
	s.allowedPaths = maps.Clone(s.config.validPaths)
	maps.Copy(s.allowedPaths, s.additionalPaths)
end:
	return s.allowedPaths
}

// ReloadConfig reloads the server configuration from disk and updates allowed paths.
// This allows for dynamic reconfiguration without restarting the server.
func (s *MCPServer) ReloadConfig() (err error) {
	var config *Config
	var allowedPaths map[string]struct{}

	// Reload config using current additional paths (empty for reload)
	// TODO: Should only post stay false, or be as defined when scout-mcp was loaded?
	config, err = s.loadConfig(Opts{OnlyMode: false})
	if err != nil {
		goto end
	}

	s.config = config

	logger.Info("Configuration reloaded")
	for dir := range allowedPaths {
		logger.Info("Allowed path", "path", dir)
	}

end:
	return err
}

// loadConfig loads the MCP server configuration from file and command-line options.
// It handles OnlyMode which ignores config files and uses only command-line paths.
func (s *MCPServer) loadConfig(opts Opts) (config *Config, err error) {
	var configFile *os.File
	var fileData []byte
	var configPath string
	var allPaths []string

	configPath, err = GetConfigPath()
	if err != nil {
		goto end
	}

	if opts.OnlyMode {
		// Use only the additional paths, ignore config file
		config = NewConfig(ConfigArgs{
			AllowedPaths:   opts.AdditionalPaths,
			Port:           ConfigPort,
			AllowedOrigins: []string{"https://claude.ai", "https://*.anthropic.com"},
		})
	} else {
		// Try to load config file
		configFile, err = os.Open(configPath)
		if err != nil {
			// If no config file and no additional paths, this is an error
			if len(opts.AdditionalPaths) == 0 {
				err = fmt.Errorf("no configuration file found and no paths specified")
				goto end
			}
			// Create minimal config with just the additional paths
			config = NewConfig(ConfigArgs{
				AllowedPaths: opts.AdditionalPaths,
				Port:         ConfigPort,
			})
		} else {
			defer mustClose(configFile)

			fileData, err = io.ReadAll(configFile)
			if err != nil {
				goto end
			}

			err = json.Unmarshal(fileData, &config)
			if err != nil {
				goto end
			}

			config.Reset()

			// Combine config paths with additional paths
			allPaths = make([]string, 0, len(config.AllowedPaths())+len(opts.AdditionalPaths))
			allPaths = append(allPaths, config.AllowedPaths()...)
			allPaths = append(allPaths, opts.AdditionalPaths...)
			config.SetAllowedPaths(allPaths)
		}
	}

	// Check if we have any paths at all
	if len(config.AllowedPaths()) == 0 {
		err = fmt.Errorf("no allowed paths specified in config file or command line")
		goto end
	}
	if opts.Stdin == nil {
		opts.Stdin = os.Stdin
	}
	s.Stdin = opts.Stdin
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	s.StdOut = opts.Stdout

	// Validate and normalize paths
	err = config.Validate()
	if err != nil {
		goto end
	}

end:
	return config, err
}

// ConfigSetter is implemented by types that can receive configuration updates.
// Tools implement this interface to be notified when server configuration changes.
type ConfigSetter interface {
	SetConfig(*Config)
}

// registerTools registers all available MCP tools with the server.
// Each tool receives the current configuration and is added to the MCP server.
func (s *MCPServer) registerTools() (err error) {
	for _, t := range mcputil.RegisteredTools() {
		t.SetConfig(s.config)
		err = s.mcpServer.AddTool(t)
		if err != nil {
			// TODO: Make this more robust
			panic(err)
		}
	}
	return err
}
