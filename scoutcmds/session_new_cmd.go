package scoutcmds

import (
	"context"

	"github.com/mikeschinkel/scout-mcp/cliutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

type SessionNewCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&SessionNewCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "new",
			Usage:       "scout session new",
			Description: "Create a new session token",
		}),
	}, &SessionCmd{})
}

func (c *SessionNewCmd) Handle(ctx context.Context, config cliutil.Config, args []string) error {
	var err error
	var session *mcputil.Session

	session = mcputil.NewSession()
	err = session.Initialize()
	if err != nil {
		goto end
	}

	cliutil.Printf("Session token: %s\n", session.Token)
	cliutil.Printf("Expires: %s\n", session.ExpiresAt.Format("2006-01-02 15:04:05"))

end:
	return err
}
