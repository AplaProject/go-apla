package cmd

import (
	"fmt"
	"os"

	"github.com/GenesisKernel/go-genesis/packages/daylight"
	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                "go-genesis",
	Short:              "Genesis application",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		daylight.Start()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
