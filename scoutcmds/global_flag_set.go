package scoutcmds

import (
	"github.com/mikeschinkel/scout-mcp/cliutil"
)

// GlobalFlagSet defines global flags available to all commands
var GlobalFlagSet = &cliutil.FlagSet{
	Name: "global",
	FlagDefs: []cliutil.FlagDef{
		{
			Name:    "config",
			Default: "",
			Usage:   "Config file path",
			String:  cfg.ConfigPath,
		},
		{
			Name:    "verbose",
			Default: false,
			Usage:   "Enable verbose logging",
			Bool:    cfg.Verbose,
		},
	},
}
