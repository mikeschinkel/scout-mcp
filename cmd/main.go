package main

import (
	"os"

	"github.com/mikeschinkel/scout-mcp"
)

func main() {
	var err error

	err = scout.RunMain(scout.RunArgs{
		Args:   os.Args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	})
	if err != nil {
		os.Exit(1)
	}
}
