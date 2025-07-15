package scout

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var allowedPaths = make(map[string]struct{})

type MCPServer struct {
	config    *Config
	mcpServer mcputil.Server
}

func NewMCPServer(opts Opts) (s *MCPServer, err error) {

	s = &MCPServer{}

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
	})

	// Register tools
	err = s.registerTools()
	if err != nil {
		goto end
	}

end:
	return s, err
}

func (s *MCPServer) StartMCP() (err error) {
	logger.Info("MCP server starting with stdio transport")
	logger.Info("Allowed directories:")
	for dir := range allowedPaths {
		logger.Info("Directory", "path", dir)
	}

	// Start the stdio server (blocking)
	err = s.mcpServer.ServeStdio()

	return err
}

func (s *MCPServer) AllowedPaths() map[string]struct{} {
	return allowedPaths
}

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

	// Validate and normalize paths
	err = config.Validate()
	if err != nil {
		goto end
	}

end:
	return config, err
}

type ConfigSetter interface {
	SetConfig(*Config)
}

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

func StartMCP(opts Opts) (err error) {
	var mcpServer *MCPServer

	mcpServer, err = NewMCPServer(opts)
	if err != nil {
		goto end
	}

	err = mcpServer.StartMCP()

end:
	return err
}
