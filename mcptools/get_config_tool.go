package mcptools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*GetConfigTool)(nil)

func init() {
	mcputil.RegisterTool(&GetConfigTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "get_config",
			Description: "Get current Scout MCP server configuration including allowed paths and settings",
			Properties: []mcputil.Property{
				mcputil.Bool("show_paths", "Show detailed allowed paths information"),
				mcputil.Bool("show_relative", "Show paths relative to home directory when possible"),
			},
		}),
	})
}

type GetConfigTool struct {
	*toolBase
}

func (t *GetConfigTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var showPaths bool
	var showRelative bool
	var config ConfigInfo

	logger.Info("Tool called", "tool", "get_config")

	showPaths = req.GetBool("show_paths", true)
	showRelative = req.GetBool("show_relative", true)

	logger.Info("Tool arguments parsed", "tool", "get_config", "show_paths", showPaths, "show_relative", showRelative)

	config, err = t.getConfigInfo(showPaths, showRelative)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	result = mcputil.NewToolResultJSON(config)

end:
	return result, err
}

type ConfigInfo struct {
	ServerName string `json:"server_name"`
	//Version        string   `json:"version"`
	AllowedPaths   []string `json:"allowed_paths"`
	AllowedOrigins []string `json:"allowed_origins"`
	PathCount      int      `json:"path_count"`
	ConfigFilePath string   `json:"config_file_path,omitempty"`
	HomeDir        string   `json:"home_directory,omitempty"`
	ServerPort     string   `json:"server_port,omitempty"`

	Summary string `json:"summary"`
}

func (t *GetConfigTool) getConfigInfo(showPaths, showRelative bool) (info ConfigInfo, err error) {
	var homeDir string
	var configFilePath string
	var allowedPaths []string
	var displayPaths []string

	// Get home directory for relative path display
	homeDir, _ = os.UserHomeDir()

	// Get config file path
	configFilePath, _ = getConfigPath()

	// Get allowed paths from the config
	allowedPaths = t.config.AllowedPaths()

	// Prepare display paths
	if showPaths {
		displayPaths = make([]string, len(allowedPaths))
		for i, path := range allowedPaths {
			displayPaths[i] = makeRelativeToHome(true, path, homeDir)
		}
	}

	cfg := t.Config()

	info = ConfigInfo{
		ServerName: "Scout MCP Server",
		//Version:      "0.0.1", // You could get this from const.go
		AllowedPaths:   displayPaths,
		AllowedOrigins: cfg.AllowedOrigins(),
		PathCount:      len(allowedPaths),
		ConfigFilePath: configFilePath,
		HomeDir:        homeDir,
		ServerPort:     cfg.ServerPort(),
		Summary: fmt.Sprintf("Scout MCP server with %d allowed director%s",
			len(allowedPaths),
			pluralize(len(allowedPaths), "y", "ies")),
	}

	if showRelative && homeDir != "" {
		info.HomeDir = "~"
	}

	info.ConfigFilePath = makeRelativeToHome(showRelative, configFilePath, homeDir)

	//end:
	return info, err
}

func makeRelativeToHome(showRelative bool, path, homeDir string) string {
	if !showRelative || homeDir == "" {
		return path
	}
	rel, err := filepath.Rel(homeDir, path)
	if err == nil {
		return "~/" + rel
	}
	return path
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "scout-mcp", "scout-mcp.json"), nil
}
