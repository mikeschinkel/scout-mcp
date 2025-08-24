package scoutcmds

import (
	"context"

	"github.com/mikeschinkel/scout-mcp/cliutil"
)

type HelpCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&HelpCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "help",
			Usage:       "scout help",
			Description: "Show help for scout commands",
		}),
	})
}

func (c *HelpCmd) Handle(ctx context.Context, config cliutil.Config, args []string) error {
	// Use the same help function that 'scout' (with no args) uses
	// This ensures consistency between 'scout' and 'scout help'
	return cliutil.ShowMainHelp()
}
