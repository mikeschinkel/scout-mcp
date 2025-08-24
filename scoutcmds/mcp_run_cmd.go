package scoutcmds

import (
	"context"
	"os"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/cliutil"
)

var MCPFlagSet = &cliutil.FlagSet{
	Name: "mcp",
	FlagDefs: []cliutil.FlagDef{
		{
			Name:    "only",
			Default: false,
			Usage:   "Use only specified paths (ignore config)",
			Bool:    cfg.OnlyMode,
		},
	},
}

type MCPRunCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&MCPRunCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "run",
			Usage:       "scout mcp run [--only] [paths...]",
			Description: "Start Scout MCP server",
			FlagSets:    []*cliutil.FlagSet{MCPFlagSet},
		}),
	}, &MCPCmd{})
}

func (c *MCPRunCmd) Handle(ctx context.Context, config cliutil.Config, args []string) error {
	var err error
	var opts *scout.Opts
	var server *scout.MCPServer

	opts, err = convertConfig(config, args)
	if err != nil {
		goto end
	}

	server, err = scout.NewMCPServer(*opts)
	if err != nil {
		goto end
	}

	fprintf(os.Stderr, "%s running...\n[Press Ctrl-C to terminate]", scout.ServerName)
	err = server.StartMCP(ctx)

end:
	return err
}
