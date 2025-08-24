package main

import (
	"context"
	"os"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/cliutil"
	"github.com/mikeschinkel/scout-mcp/scoutcmds"
)

func main() {
	err := scout.RunMain(context.Background(), scout.RunArgs{
		Args:           os.Args,
		MCPReader:      os.Stdin,
		MCPWriter:      os.Stdout,
		CLIWriter:      cliutil.NewOutputWriter(),
		ConfigProvider: scoutcmds.NewConfigProvider(),
	})
	if err != nil {
		os.Exit(1)
	}
}
