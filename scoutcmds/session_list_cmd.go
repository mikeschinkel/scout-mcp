package scoutcmds

import (
	"context"

	"github.com/mikeschinkel/scout-mcp/cliutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

type SessionListCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&SessionListCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "list",
			Usage:       "scout session list",
			Description: "List all active sessions",
		}),
	}, &SessionCmd{})
}

func (c *SessionListCmd) Handle(ctx context.Context, config cliutil.Config, args []string) error {
	var err error
	var sessions []mcputil.Session

	sessions = mcputil.ListSessions()

	if len(sessions) == 0 {
		cliutil.Printf("No active sessions\n")
		goto end
	}

	cliutil.Printf("Active sessions:\n")
	for _, session := range sessions {
		cliutil.Printf("  %s (expires: %s)\n", session.Token, session.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

end:
	return err
}
