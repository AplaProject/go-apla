package modes

import (
	"io/ioutil"
	"path/filepath"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/install"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	log "github.com/sirupsen/logrus"
)

// InitBlockchain creates exemplar of Blochain mode
func InitBlockchain(config *conf.SavedConfig) *Blockchain {
	mode := &Blockchain{
		SavedConfig: config,
	}

	return mode
}

// Blockchain represent implementation to run node as blockchain
type Blockchain struct {
	*conf.SavedConfig
}

// Start Implement NodeMode interface
func (mode *Blockchain) Start(exitFunc func(int), gormFunc func(conf.DBConfig)) {

	if *conf.GenerateFirstBlock {
		if err := install.GenerateFirstBlock(); err != nil {
			log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("GenerateFirstBlock")
			exitFunc(1)
		}
	}

	if conf.Installed {
		if conf.Config.KeyID == 0 {
			key, err := parser.GetKeyIDFromPrivateKey()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("Unable to get KeyID")
				exitFunc(1)
			}
			conf.Config.KeyID = key
			if err := conf.SaveConfig(); err != nil {
				log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("Error writing config file")
				exitFunc(1)
			}
		}
		gormFunc(conf.Config.DB)
	}

	// database rollback to the specified block
	if *conf.RollbackToBlockID > 0 {

		err := syspar.SysUpdate(nil)
		if err != nil {
			log.WithError(err).Error("can't read system parameters")
		}

		log.WithFields(log.Fields{"block_id": *conf.RollbackToBlockID}).Info("Rollbacking to block ID")

		if err := rollbackToBlock(*conf.RollbackToBlockID); err != nil {
			log.WithError(err).Error("Rollback error")
		} else {
			log.WithFields(log.Fields{"block_id": *conf.RollbackToBlockID}).Info("Rollback is ok")
		}

		exitFunc(0)
	}
}

func rollbackToBlock(blockID int64) error {
	if err := smart.LoadContracts(nil); err != nil {
		return err
	}
	parser := new(parser.Parser)
	err := parser.RollbackToBlockID(*conf.RollbackToBlockID)
	if err != nil {
		return err
	}

	// block id = 1, is a special case for full rollback
	if blockID != 1 {
		return nil
	}

	// check blocks related tables
	startData := map[string]int64{"1_menu": 1, "1_pages": 1, "1_contracts": 26, "1_parameters": 11, "1_keys": 1, "1_tables": 8, "stop_daemons": 1, "queue_blocks": 9999999, "system_tables": 1, "system_parameters": 27, "system_states": 1, "install": 1, "queue_tx": 9999999, "log_transactions": 1, "transactions_status": 9999999, "block_chain": 1, "info_block": 1, "confirmations": 9999999, "transactions": 9999999}
	warn := 0
	for table := range startData {
		count, err := model.GetRecordsCountTx(nil, table)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting record count")
			return err
		}
		if count > 0 && count > startData[table] {
			log.WithFields(log.Fields{"count": count, "start_data": startData[table], "table": table}).Warn("record count in table is larger then start")
			warn++
		} else {
			log.WithFields(log.Fields{"count": count, "start_data": startData[table], "table": table}).Info("record count in table is ok")
		}
	}

	if warn == 0 {
		rbFile := filepath.Join(conf.Config.WorkDir, consts.RollbackResultFilename)
		ioutil.WriteFile(rbFile, []byte("1"), 0644)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.WritingFile, "path": rbFile}).Error("rollback result flag")
			return err
		}
	}
	return nil
}
