package cmd

import (
	"flag"
	"os"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var blockID int64
var dbHost string
var dbPort int
var dbName string
var dbUser string

// rollbackCmd represents the rollback command
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollbacks blockchain to blockID",
	Run: func(cmd *cobra.Command, args []string) {
		dbPassword := os.Getenv("DB_PASSWORD")
		flag.Parse()
		if err := model.GormInit(dbHost, dbPort, dbUser, dbPassword, dbName); err != nil {
			log.WithError(err).Fatal("init db")
			return
		}
		if err := syspar.SysUpdate(nil); err != nil {
			log.WithError(err).Error("can't read system parameters")
		}
		if err := smart.LoadContracts(nil); err != nil {
			log.WithError(err).Fatal("loading contracts")
			return
		}
		parser := new(parser.Parser)
		err := parser.RollbackToBlockID(blockID)
		if err != nil {
			log.WithError(err).Fatal("rollback to block id")
			return
		}

		// block id = 1, is a special case for full rollback
		if blockID != 1 {
			log.Info("Not full rollback, finishing work without checking")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)

	rollbackCmd.Flags().Int64Var(&blockID, "blockID", 1, "blockID to rollback")
	rollbackCmd.Flags().IntVar(&dbPort, "dbPort", 5432, "genesis database port to rollback")
	rollbackCmd.Flags().StringVar(&dbHost, "dbHost", "localhost", "genesis database host to rollback")
	rollbackCmd.Flags().StringVar(&dbName, "dbName", "genesis", "genesis database name")
	rollbackCmd.Flags().StringVar(&dbUser, "dbUser", "genesis", "genesis database username")

}
