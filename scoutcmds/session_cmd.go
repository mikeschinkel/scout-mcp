package scoutcmds

import (
	"github.com/mikeschinkel/scout-mcp/cliutil"
)

type SessionCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&SessionCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "session",
			Usage:       "scout session [subcommand]",
			Description: "Session management operations",
			DelegateTo:  &SessionNewCmd{}, // Default to new command
		}),
	})
}
