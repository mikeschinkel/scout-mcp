package scoutcmds

import (
	"context"

	"github.com/mikeschinkel/scout-mcp/cliutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var sessionTokenArg = new(string)

type SessionClearCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&SessionClearCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "clear",
			Usage:       "scout session clear <token|all>",
			Description: "Clear specific session or all sessions",
			ArgDefs: []*cliutil.ArgDef{
				{
					Name:     "token",
					Usage:    "Session token to clear or 'all' for all sessions",
					Required: true,
					String:   sessionTokenArg,
				},
			},
		}),
	}, &SessionCmd{})
}

func (c *SessionClearCmd) Handle(ctx context.Context, config cliutil.Config, args []string) (err error) {
	switch *sessionTokenArg {
	case "all":
		err = mcputil.ClearSessions(mcputil.AllSessions)
		if err != nil {
			goto end
		}
		cliutil.Printf("All sessions cleared\n")
	default:
		if mcputil.ClearSession(*sessionTokenArg) {
			cliutil.Printf("Session %s cleared\n", *sessionTokenArg)
			goto end
		}
		cliutil.Printf("Session not found: %s\n", *sessionTokenArg)
	}
end:
	return err
}
