package cmd

import (
	"github.com/GenesisKernel/go-genesis/packages/daylight"
	"github.com/spf13/cobra"
)

// startCmd is alias to root command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starting node",
	Run: func(cmd *cobra.Command, args []string) {
		daylight.Start()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
