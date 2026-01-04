package main

import (
	"os"

	"github.com/Kizsoft-Solution-Limited/uniroute/cmd/cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
