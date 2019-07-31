// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"os"
	"path/filepath"

	"github.com/AplaProject/go-apla/packages/conf"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-apla",
	Short: "Apla application",
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
		versionCmd,
	)

	// This flags are visible for all child commands
	rootCmd.PersistentFlags().StringVar(&conf.Config.ConfigPath, "config", defautConfigPath(), "filepath to config.toml")
}

// Execute executes rootCmd command.
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

	return filepath.Join(p, "apla-data", "config.toml")
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
