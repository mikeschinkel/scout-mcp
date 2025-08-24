package scoutcmds

import (
	"github.com/mikeschinkel/scout-mcp/cliutil"
)

type MCPCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&MCPCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "mcp",
			Usage:       "scout mcp [subcommand]",
			Description: "MCP server operations",
			DelegateTo:  &MCPRunCmd{}, // Default to run command
		}),
	})
}
