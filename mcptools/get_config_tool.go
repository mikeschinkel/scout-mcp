package mcptools

import (
	"context"
	"fmt"
	"os"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*GetConfigTool)(nil)

func init() {
	mcputil.RegisterTool(&GetConfigTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "get_config",
			Description: "Get current Scout MCP server configuration including allowed paths and settings",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
			},
		}),
	})
}

type GetConfigTool struct {
	*mcputil.ToolBase
}

func (t *GetConfigTool) Handle(_ context.Context, _ mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var config ConfigInfo

	logger.Info("Tool called", "tool", "get_config")

	logger.Info("Tool arguments parsed", "tool", "get_config")

	config, err = t.getConfigInfo(t.Config())
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	logger.Info("Tool completed", "tool", "get_config", "success", true)
	result = mcputil.NewToolResultJSON(config)

end:
	return result, err
}

type ConfigInfo struct {
	ServerName     string   `json:"server_name"`
	AllowedPaths   []string `json:"allowed_paths"`
	AllowedOrigins []string `json:"allowed_origins"`
	PathCount      int      `json:"path_count"`
	ConfigFilePath string   `json:"config_file_path"`
	HomeDirectory  string   `json:"home_directory"`
	ServerPort     string   `json:"server_port"`
	Summary        string   `json:"summary"`
}

func (t *GetConfigTool) getConfigInfo(cfg mcputil.Config) (info ConfigInfo, err error) {
	var allowedPaths []string
	var homeDir string
	var configPath string

	// Get allowed paths from the config
	allowedPaths = cfg.AllowedPaths()

	// Get home directory for relative path display
	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	// Get config path via the interface
	if t.Config().Path() != "" {
		configPath, err = makeRelativePath(t.Config().Path(), homeDir)
		if err != nil {
			goto end
		}
	} else {
		// For tests or when no config file is used
		configPath = "(no config file)"
	}

	info = ConfigInfo{
		ServerName:     cfg.ServerName(),
		ServerPort:     cfg.ServerPort(),
		AllowedPaths:   allowedPaths,
		AllowedOrigins: cfg.AllowedOrigins(),
		PathCount:      len(allowedPaths),
		ConfigFilePath: configPath,
		HomeDirectory:  homeDir,
		Summary: fmt.Sprintf("Scout MCP server with %d allowed director%s",
			len(allowedPaths),
			func() string {
				if len(allowedPaths) == 1 {
					return "y"
				} else {
					return "ies"
				}
			}()),
	}

end:
	return info, err
}
