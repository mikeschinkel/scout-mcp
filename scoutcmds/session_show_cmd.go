package scoutcmds

import (
	"context"

	"github.com/mikeschinkel/scout-mcp/cliutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var showTokenArg = new(string)

type SessionShowCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&SessionShowCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "show",
			Usage:       "scout session show <token>",
			Description: "Show session details",
			ArgDefs: []*cliutil.ArgDef{
				{
					Name:     "token",
					Usage:    "Session token to show details for",
					Required: true,
					String:   showTokenArg,
				},
			},
		}),
	}, &SessionCmd{})
}

func (c *SessionShowCmd) Handle(ctx context.Context, config cliutil.Config, args []string) error {
	var err error
	var session *mcputil.Session
	var exists bool

	session, exists = mcputil.GetSession(*showTokenArg)
	if !exists {
		cliutil.Printf("Session not found: %s\n", *showTokenArg)
		goto end
	}

	cliutil.Printf("Session: %s\n", session.Token)
	cliutil.Printf("Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
	cliutil.Printf("Expires: %s\n", session.ExpiresAt.Format("2006-01-02 15:04:05"))
	if !session.LastUsed.IsZero() {
		cliutil.Printf("Last used: %s\n", session.LastUsed.Format("2006-01-02 15:04:05"))
	} else {
		cliutil.Printf("Last used: never\n")
	}

end:
	return err
}
