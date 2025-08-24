package scoutcmds

import (
	"github.com/mikeschinkel/scout-mcp/cliutil"
)

type ToolCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&ToolCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "tool",
			Usage:       "scout tool [subcommand]",
			Description: "Tool operations",
			DelegateTo:  &ToolListCmd{}, // Default to list command
		}),
	})
}
