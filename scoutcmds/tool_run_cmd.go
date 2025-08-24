package scoutcmds

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/scout-mcp/cliutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var toolNameArg = new(string)

type ToolRunCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&ToolRunCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "run",
			Usage:       "scout tool run <tool-name> [args...]",
			Description: "Execute a specific MCP tool",
			ArgDefs: []*cliutil.ArgDef{
				{
					Name:     "tool-name",
					Usage:    "Name of tool to run",
					Required: true,
					String:   toolNameArg,
				},
			},
		}),
	}, &ToolCmd{})
}

func (c *ToolRunCmd) Handle(ctx context.Context, config cliutil.Config, args []string) error {
	var err error
	var tool mcputil.Tool
	var exists bool

	// Get the tool
	tools := mcputil.RegisteredToolsMap()
	tool, exists = tools[*toolNameArg]
	if !exists {
		err = fmt.Errorf("unknown tool: %s", *toolNameArg)
		goto end
	}

	// For now, just display tool information
	// TODO: In the future, implement actual tool execution with proper request building
	cliutil.Printf("Tool: %s\n", tool.Options().Name)
	cliutil.Printf("Description: %s\n", tool.Options().Description)
	cliutil.Printf("Properties: %d\n", len(tool.Options().Properties))

	cliutil.Printf("\nNote: Tool execution from CLI is not yet implemented.\n")
	cliutil.Printf("Use this command with MCP server mode: scout mcp\n")

end:
	return err
}
