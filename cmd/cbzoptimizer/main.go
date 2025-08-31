package main

import (
	"github.com/danielkitchener/CBZOptimizer/v2/cmd/cbzoptimizer/commands"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	commands.SetVersionInfo(version, commit, date)

	commands.Execute()
}
