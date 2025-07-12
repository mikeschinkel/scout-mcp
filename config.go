package scout

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	WhitelistedPaths []string `json:"whitelisted_paths"`
	Port             string   `json:"port"`
	AllowedOrigins   []string `json:"allowed_origins"`
}

type ConfigArgs struct {
	InitialPath string
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

func CreateDefaultConfig(args ConfigArgs) (err error) {
	var homeDir string
	var config Config
	var configData []byte
	var configFile *os.File
	var configPath string
	var configDir string
	var whitelistedPaths []string

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
		whitelistedPaths = []string{args.InitialPath}
	} else {
		// Create empty config that user must edit
		whitelistedPaths = []string{}
	}

	config = Config{
		WhitelistedPaths: whitelistedPaths,
		Port:             ConfigPort,
		AllowedOrigins:   []string{"https://claude.ai", "https://*.anthropic.com"},
	}

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
		logger.Info("Initial whitelisted directory", "path", args.InitialPath)
	} else {
		logger.Info("Empty config created - you must edit the config file to add whitelisted paths")
	}

end:
	return err
}
