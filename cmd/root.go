package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"os"
	"path/filepath"

	"github.com/GenesisKernel/go-genesis/packages/conf"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-genesis",
	Short: "Genesis application",
}

func init() {
	rootCmd.AddCommand(
		generateFirstBlockCmd,
		generateKeysCmd,
		initDatabaseCmd,
		rollbackCmd,
		startCmd,
		configCmd,
		stopNetworkCmd,
	)

	// This flags are visible for all child commands
	rootCmd.PersistentFlags().StringVar(&conf.Config.ConfigPath, "config", defautConfigPath(), "filepath to config.toml")
}

// This is called by main.main(). It only needs to happen once to the rootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Executing root command")
	}
}

func defautConfigPath() string {
	p, err := os.Getwd()
	if err != nil {
		log.WithError(err).Fatal("getting cur wd")
	}

	return filepath.Join(p, "genesis-data", "config.toml")
}

// Load the configuration from file
func loadConfig(cmd *cobra.Command, args []string) {
	err := conf.LoadConfig(conf.Config.ConfigPath)
	if err != nil {
		log.WithError(err).Fatal("Loading config")
	}
}

func loadConfigWKey(cmd *cobra.Command, args []string) {
	err := conf.LoadConfig(conf.Config.ConfigPath)
	if err != nil {
		log.WithError(err).Fatal("Loading config")
	}

	err = conf.FillRuntimeKey()
	if err != nil {
		log.WithError(err).Fatal("Filling keys")
	}
}
