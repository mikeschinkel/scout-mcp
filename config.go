package scout

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var allowedOrigins = []string{
	"https://claude.ai",
	"https://*.anthropic.com",
}

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
		AllowedPaths:   allowedPaths,
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

// Config struct with private fields + embedded JSONConfig
type Config struct {
	JSONConfig                     // Embedded for JSON operations
	validPaths map[string]struct{} // Private runtime index
	path       string
}

func (c *Config) ServerName() string {
	return ServerName
}

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

// JSONConfig for serialization (exported fields)
type JSONConfig struct {
	AllowedPaths   []string `json:"allowed_paths"`
	Port           string   `json:"port"`
	AllowedOrigins []string `json:"allowed_origins"`
}

type ConfigArgs struct {
	InitialPath    string
	AllowedPaths   []string
	Port           string
	AllowedOrigins []string
}

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

func (c *Config) Initialize() (err error) {
	c.path, err = GetConfigPath()
	return err
}
func (c *Config) Path() string {
	return c.path
}
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

func (c *Config) SetAllowedPaths(paths []string) {
	c.JSONConfig.AllowedPaths = slices.Clone(paths)
}

func (c *Config) ServerPort() string {
	return c.JSONConfig.Port
}

func (c *Config) AllowedOrigins() []string {
	return c.JSONConfig.AllowedOrigins
}

func (c *Config) Reset() {
	jc := c.JSONConfig
	if jc.AllowedPaths == nil {
		c.JSONConfig.AllowedPaths = []string{}
	}
	c.validPaths = make(map[string]struct{})
	if jc.AllowedOrigins == nil {
		c.JSONConfig.AllowedOrigins = allowedOrigins
	}
}

func (c *Config) IsAllowedPath(targetPath string) (allowed bool, err error) {
	var absPath string

	absPath, err = filepath.Abs(targetPath)
	if err != nil {
		goto end
	}

	for path := range c.validPaths {
		if strings.HasPrefix(absPath, path) {
			allowed = true
			goto end
		}
	}

	allowed = false

end:
	return allowed, err
}

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
