package cmd

import (
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var blockID int64

// rollbackCmd represents the rollback command
var rollbackCmd = &cobra.Command{
	Use:    "rollback",
	Short:  "Rollback blockchain to blockID",
	PreRun: loadConfigWKey,
	Run: func(cmd *cobra.Command, args []string) {
		f := utils.LockOrDie(conf.Config.LockFilePath)
		defer f.Unlock()

		if err := model.GormInit(
			conf.Config.DB.Host,
			conf.Config.DB.Port,
			conf.Config.DB.User,
			conf.Config.DB.Password,
			conf.Config.DB.Name,
		); err != nil {
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
	rollbackCmd.Flags().Int64Var(&blockID, "blockId", 1, "blockID to rollback")
	rollbackCmd.MarkFlagRequired("blockId")
}
