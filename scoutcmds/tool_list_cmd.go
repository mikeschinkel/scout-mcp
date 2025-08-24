package scoutcmds

import (
	"context"

	"github.com/mikeschinkel/scout-mcp/cliutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

type ToolListCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&ToolListCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "list",
			Usage:       "scout tool list",
			Description: "List available tools",
		}),
	}, &ToolCmd{})
}

func (c *ToolListCmd) Handle(ctx context.Context, config cliutil.Config, args []string) error {
	var err error
	var tools map[string]mcputil.Tool

	tools = mcputil.RegisteredToolsMap()

	cliutil.Printf("Available tools:\n")
	for name, tool := range tools {
		cliutil.Printf("  %-20s %s\n", name, tool.Options().Description)
	}

	return err
}
