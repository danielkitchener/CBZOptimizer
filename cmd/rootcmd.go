package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "cbzconverter",
	Short: "Convert CBZ files using a specified converter",
}

func init() {
	viper.SetConfigName("CBZOptimizer")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("CBZ")
	viper.AutomaticEnv()
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
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
