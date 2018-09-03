package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GenesisKernel/go-genesis/packages/consts"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(consts.VERSION)
	},
}
