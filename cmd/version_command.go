package cmd

import (
	"fmt"
	"github.com/belphemur/CBZOptimizer/meta"
	"github.com/spf13/cobra"
)

func init() {
	command := &cobra.Command{
		Use:   "version",
		Short: "Print the version of the application",
		Long:  "Print the version of the application",
		Run:   VersionCommand,
	}
	AddCommand(command)
}

func VersionCommand(_ *cobra.Command, _ []string) {
	fmt.Printf("CBZOptimizer %s [%s] built [%s]\n", meta.Version, meta.Commit, meta.Date)
}
