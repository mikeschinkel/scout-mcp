package scoutcmds

import (
	"context"
	"path/filepath"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/cliutil"
)

var initPathArg = new(string)

type InitCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&InitCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "init",
			Usage:       "scout init [path]",
			Description: "Create or update Scout configuration file",
			ArgDefs: []*cliutil.ArgDef{
				{
					Name:     "path",
					Usage:    "Directory path to add to allowed paths",
					Required: false,
					String:   initPathArg,
				},
			},
		}),
	})
}

func (c *InitCmd) Handle(ctx context.Context, config cliutil.Config, args []string) error {
	var err error
	var configPath string
	var scoutArgs scout.Args

	// Get config file path
	configPath, err = scout.GetConfigPath()
	if err != nil {
		goto end
	}

	if cfg.ConfigPath != nil && *cfg.ConfigPath != "" {
		configPath = *cfg.ConfigPath
	}

	// Add the provided path if specified
	if initPathArg != nil && *initPathArg != "" {
		pathToAdd, err := filepath.Abs(*initPathArg)
		if err != nil {
			goto end
		}
		scoutArgs.AdditionalPaths = []string{pathToAdd}
	}

	// Create/update configuration
	err = scout.CreateDefaultConfig(scoutArgs)
	if err != nil {
		goto end
	}

	cliutil.Printf("Configuration saved to: %s\n", configPath)
	if len(scoutArgs.AdditionalPaths) > 0 {
		cliutil.Printf("Added path: %s\n", scoutArgs.AdditionalPaths[0])
	}

end:
	return err
}
