package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mikeschinkel/scout-mcp"
)

func main() {
	var err error

	_, _ = fmt.Fprintf(os.Stderr, "%s running...\n[Press Ctrl-C to terminate]", scout.ServerName)
	err = scout.RunMain(context.Background(), scout.RunArgs{
		Args:   os.Args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	})
	if err != nil {
		os.Exit(1)
	}
}
