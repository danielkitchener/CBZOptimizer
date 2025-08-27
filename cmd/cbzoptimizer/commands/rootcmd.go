package commands

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime"
)

var rootCmd = &cobra.Command{
	Use:   "cbzconverter",
	Short: "Convert CBZ files using a specified converter",
}

func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

func getPath() string {
	return filepath.Join(map[string]string{
		"windows": filepath.Join(os.Getenv("APPDATA")),
		"darwin":  filepath.Join(os.Getenv("HOME"), ".config"),
		"linux":   filepath.Join(os.Getenv("HOME"), ".config"),
	}[runtime.GOOS], "CBZOptimizer")
}

func init() {
	configFolder := getPath()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configFolder)
	viper.SetEnvPrefix("CBZ")
	viper.AutomaticEnv()
	err := os.MkdirAll(configFolder, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err := viper.SafeWriteConfig()
			if err != nil {
				panic(fmt.Errorf("fatal error config file: %w", err))
			}
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Command execution failed")
	}
}
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
