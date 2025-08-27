package main

import (
	"github.com/belphemur/CBZOptimizer/v2/cmd/cbzoptimizer/commands"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	commands.SetVersionInfo(version, commit, date)
	commands.Execute()
}
