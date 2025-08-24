package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/cliutil"
	"github.com/mikeschinkel/scout-mcp/scoutcmds"
)

func main() {
	logger, err := scout.CreateJSONLogger()
	if err != nil {
		err = fmt.Errorf("failed to initialize logger: %v\n", err)
		goto end
	}
	err = scout.Run(context.Background(), scout.RunArgs{
		Args:           os.Args,
		Logger:         logger,
		MCPReader:      os.Stdin,
		MCPWriter:      os.Stdout,
		CLIWriter:      cliutil.NewOutputWriter(),
		ConfigProvider: scoutcmds.NewConfigProvider(),
	})
end:
	if err != nil {
		os.Exit(1)
	}
}
