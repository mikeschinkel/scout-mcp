package scout

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// allowedOrigins contains the default list of allowed origins for MCP connections.
// These origins are always included in the server configuration.
var allowedOrigins = []string{
	"https://claude.ai",
	"https://*.anthropic.com",
}

// GetConfigPath returns the full path to the Scout-MCP configuration file
// in the user's home directory (~/.config/scout-mcp/scout-mcp.json).
func GetConfigPath() (configPath string, err error) {
	var homeDir string
	var configDir string

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	configDir = filepath.Join(homeDir, ConfigBaseDirName, ConfigDirName)
	configPath = filepath.Join(configDir, ConfigFileName)

end:
	return configPath, err
}

// CreateDefaultConfig creates a new configuration file with default settings
// and the specified allowed paths. The config file is written to the user's
// Scout-MCP configuration directory.
func CreateDefaultConfig(args Args) (err error) {
	var homeDir string
	var config *Config
	var configData []byte
	var configFile *os.File
	var configPath string
	var configDir string
	var allowedPaths []string

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	configDir = filepath.Join(homeDir, ConfigBaseDirName, ConfigDirName)
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		goto end
	}

	configPath = filepath.Join(configDir, ConfigFileName)

	if args.InitialPath != "" {
		// Validate the provided path
		err = validatePath(args.InitialPath)
		if err != nil {
			goto end
		}
		allowedPaths = []string{args.InitialPath}
	} else {
		// Create empty config that user must edit
		allowedPaths = []string{}
	}

	config = NewConfig(ConfigArgs{
		AllowedPaths:   append(allowedPaths, "/tmp"),
		Port:           ConfigPort,
		AllowedOrigins: allowedOrigins,
	})

	configData, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		goto end
	}

	configFile, err = os.Create(configPath)
	if err != nil {
		goto end
	}
	defer mustClose(configFile)

	_, err = configFile.Write(configData)
	if err != nil {
		goto end
	}

	logger.Info("Created config file", "path", configPath)
	if args.InitialPath != "" {
		logger.Info("Initial allowed directory", "path", args.InitialPath)
		goto end
	}

	logger.Info("Empty config created - you must edit the config file to add allowed paths", "path", configPath)

end:
	return err
}

var _ mcputil.Config = (*Config)(nil)

// Config represents the Scout-MCP server configuration, containing
// allowed paths, port settings, and runtime validation state.
// It embeds JSONConfig for serialization while maintaining private runtime data.
type Config struct {
	JSONConfig                     // Embedded for JSON operations
	validPaths map[string]struct{} // Private runtime index
	path       string
}

// ServerName returns the name of the MCP server.
func (c *Config) ServerName() string {
	return ServerName
}

// ToMap converts the Config to a map[string]any representation
// for JSON serialization and API responses.
func (c *Config) ToMap() (m map[string]any, err error) {
	var b []byte
	b, err = json.Marshal(c)
	if err != nil {
		goto end
	}
	err = json.Unmarshal(b, &m)
	if err != nil {
		goto end
	}
end:
	return m, err
}

// JSONConfig contains the serializable configuration fields that are
// written to and read from the configuration file.
type JSONConfig struct {
	AllowedPaths   []string `json:"allowed_paths"`
	Port           string   `json:"port"`
	AllowedOrigins []string `json:"allowed_origins"`
}

// ConfigArgs contains the arguments needed to create a new Config instance,
// including initial paths, port, and allowed origins.
type ConfigArgs struct {
	InitialPath    string
	AllowedPaths   []string
	Port           string
	AllowedOrigins []string
}

// NewConfig creates a new Config instance with the provided arguments.
// It initializes both the JSON-serializable fields and runtime validation state.
func NewConfig(args ConfigArgs) *Config {
	c := &Config{
		JSONConfig: JSONConfig{
			AllowedPaths:   args.AllowedPaths,
			Port:           args.Port,
			AllowedOrigins: append(args.AllowedOrigins, allowedOrigins...),
		},
	}
	c.Reset()
	return c
}

// Initialize sets up the config with the default configuration file path.
func (c *Config) Initialize() (err error) {
	c.path, err = GetConfigPath()
	return err
}

// Path returns the file path where the configuration is stored
func (c *Config) Path() string {
	return c.path
}

// AllowedPaths returns a deduplicated slice of all allowed directory paths
// AllowedPaths returns a deduplicated slice of all allowed directory paths.
func (c *Config) AllowedPaths() []string {
	index := make(map[string]struct{}, len(c.JSONConfig.AllowedPaths))
	for _, path := range c.JSONConfig.AllowedPaths {
		if path == "" {
			continue
		}
		index[path] = struct{}{}
	}
	return slices.Collect(maps.Keys(index))
}

// SetAllowedPaths updates the list of allowed paths with a clone of the provided slice.
func (c *Config) SetAllowedPaths(paths []string) {
	c.JSONConfig.AllowedPaths = slices.Clone(paths)
}

// ServerPort returns the configured server port.
func (c *Config) ServerPort() string {
	return c.JSONConfig.Port
}

// AllowedOrigins returns the list of allowed request origins.
func (c *Config) AllowedOrigins() []string {
	return c.JSONConfig.AllowedOrigins
}

// Reset initializes the config's runtime state including default paths and origins.
func (c *Config) Reset() {
	c.validPaths = make(map[string]struct{})
	c.validPaths["/tmp"] = struct{}{}
	jc := c.JSONConfig
	if jc.AllowedPaths == nil {
		c.JSONConfig.AllowedPaths = []string{"/tmp"}
	}
	if jc.AllowedOrigins == nil {
		c.JSONConfig.AllowedOrigins = allowedOrigins
	}
}

// IsAllowedPath checks if the given path is within any of the allowed directories.
// It converts the target path to absolute form and checks against all allowed paths.
func (c *Config) IsAllowedPath(targetPath string) (allowed bool) {

	// Fails one if os.Getwd() fails, so lets ignore that as that failure will be
	// caught elsewhere.
	targetPath, _ = filepath.Abs(targetPath)

	for path := range c.validPaths {
		_, err := filepath.Rel(path, targetPath)
		if err == nil {
			allowed = true
			goto end
		}
	}

end:
	return allowed
}

// Validate checks that all configured paths exist and are directories,
// building the internal validation map for runtime path checking.
func (c *Config) Validate() (err error) {
	var absPath string
	var pathInfo os.FileInfo
	var errs []error

	for _, path := range c.AllowedPaths() {
		absPath, err = filepath.Abs(path)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		pathInfo, err = os.Stat(absPath)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if !pathInfo.IsDir() {
			errs = append(errs, fmt.Errorf("allowed path is not a directory: %s", absPath))
			continue
		}

		c.validPaths[absPath] = struct{}{}
	}

	logger.Info("allowed paths", "paths", c.validPaths)

	return err
}
