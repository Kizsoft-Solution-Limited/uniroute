package main

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/Kizsoft-Solution-Limited/uniroute/cmd/cli/commands"
)

func main() {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		commands.SetVersion(strings.TrimPrefix(info.Main.Version, "v"))
	}
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
