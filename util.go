package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func mustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func validatePath(path string) (err error) {
	var absPath string
	var pathInfo os.FileInfo

	absPath, err = filepath.Abs(path)
	if err != nil {
		err = fmt.Errorf("invalid path '%s': %v", path, err)
		goto end
	}

	pathInfo, err = os.Stat(absPath)
	if err != nil {
		err = fmt.Errorf("path '%s' does not exist: %v", absPath, err)
		goto end
	}

	if !pathInfo.IsDir() {
		err = fmt.Errorf("path '%s' is not a directory", absPath)
		goto end
	}

end:
	return err
}

func parseArgs() (additionalPaths []string, opts MCPServerOpts, isInit bool, args ConfigArgs, err error) {
	var osArgs []string
	var i int
	var arg string

	osArgs = os.Args[1:] // Skip program name

	if len(osArgs) == 0 {
		return additionalPaths, opts, isInit, args, err
	}

	// Check for init command
	if osArgs[0] == "init" {
		isInit = true
		if len(osArgs) > 1 {
			args.InitialPath = osArgs[1]
		}
		return additionalPaths, opts, isInit, args, err
	}

	// Parse flags and paths
	for i = 0; i < len(osArgs); i++ {
		arg = osArgs[i]

		if arg == "--only" {
			opts.OnlyMode = true
			continue
		}

		// Validate and add path
		err = validatePath(arg)
		if err != nil {
			goto end
		}

		additionalPaths = append(additionalPaths, arg)
	}

end:
	return additionalPaths, opts, isInit, args, err
}

func showUsageError() {
	var homeDir string
	var configPath string

	homeDir, _ = os.UserHomeDir()
	if homeDir != "" {
		configPath = filepath.Join(homeDir, ConfigBaseDirName, ConfigDirName, ConfigFileName)
	} else {
		configPath = "~/" + ConfigBaseDirName + "/" + ConfigDirName + "/" + ConfigFileName
	}

	logger.Error("No whitelisted directories specified")
	logger.Info("Usage options")
	logger.Info("Add path to config file paths", "command", fmt.Sprintf("%s <path>", os.Args[0]))
	logger.Info("Use only the specified path", "command", fmt.Sprintf("%s --only <path>", os.Args[0]))
	logger.Info("Create empty config file", "command", fmt.Sprintf("%s init", os.Args[0]))
	logger.Info("Create config with custom initial path", "command", fmt.Sprintf("%s init <path>", os.Args[0]))
	logger.Info("Config file location", "path", configPath)
	logger.Info("Example commands")
	logger.Info("Example", "command", fmt.Sprintf("%s ~/MyProjects", os.Args[0]))
	logger.Info("Example", "command", fmt.Sprintf("%s --only /tmp/safe-dir", os.Args[0]))
	logger.Info("Example", "command", fmt.Sprintf("%s init ~/Code", os.Args[0]))
}
