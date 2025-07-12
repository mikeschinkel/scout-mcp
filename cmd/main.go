package main

import (
	"os"

	"github.com/mikeschinkel/scout-mcp"
)

func main() {
	var err error

	err = scout.RunMain()
	if err != nil {
		os.Exit(1)
	}
}
