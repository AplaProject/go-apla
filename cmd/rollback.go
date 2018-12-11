package cmd

import (
	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/rollback"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var blockHash string

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
		if err := smart.LoadContracts(); err != nil {
			log.WithError(err).Fatal("loading contracts")
			return
		}
		binBlockHash := converter.HexToBin(blockHash)
		err := rollback.ToBlockID(binBlockHash, nil, nil, log.WithFields(log.Fields{}))
		if err != nil {
			log.WithError(err).Fatal("rollback to block hash")
			return
		}
	},
}

func init() {
	rollbackCmd.Flags().StringVar(&blockHash, "blockHash", "", "blockHash to rollback")
	rollbackCmd.MarkFlagRequired("blockHash")
}
