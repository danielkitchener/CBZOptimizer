package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "cbzconverter",
	Short: "Convert CBZ files using a specified converter",
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
