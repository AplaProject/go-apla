package daemons

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

func Monitoring(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	infoBlock := &model.InfoBlock{}
	found, err := infoBlock.Get()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting info block")
		logError(w, fmt.Errorf("can't get info block: %s", err))
		return
	}
	if !found {
		logError(w, fmt.Errorf("can't find info block: %s", err))
		return
	}
	addKey(&buf, "info_block_id", infoBlock.BlockID)
	addKey(&buf, "info_block_hash", converter.BinToHex(infoBlock.Hash))
	addKey(&buf, "info_block_time", infoBlock.Time)
	addKey(&buf, "info_block_wallet", infoBlock.WalletID)
	addKey(&buf, "info_block_state", infoBlock.StateID)

	fullNode := &model.FullNode{}
	nodes, err := fullNode.GetAll()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting full nodes")
		logError(w, fmt.Errorf("can't get full nodes: %s", err))
		return
	}
	addKey(&buf, "full_nodes_count", len(*nodes))

	block := &model.Block{}
	found, err = block.GetMaxBlock()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting max block")
		logError(w, fmt.Errorf("can't get max block: %s", err))
		return
	}
	if !found {
		logError(w, fmt.Errorf("can't find info block"))
	}
	addKey(&buf, "last_block_id", block.ID)
	addKey(&buf, "last_block_hash", converter.BinToHex(block.Hash))
	addKey(&buf, "last_block_time", block.Time)
	addKey(&buf, "last_block_wallet", block.WalletID)
	addKey(&buf, "last_block_state", block.StateID)
	addKey(&buf, "last_block_transactions", block.Tx)

	trCount, err := model.GetTransactionCountAll()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting transaction count all")
		logError(w, fmt.Errorf("can't get transactions count: %s", err))
		return
	}
	addKey(&buf, "transactions_count", trCount)

	w.Write(buf.Bytes())
}

func addKey(buf *bytes.Buffer, key string, value interface{}) {
	line := fmt.Sprintf("%s\t%s\n", key, converter.InterfaceToStr(value))
	buf.Write([]byte(line))
}

func logError(w http.ResponseWriter, err error) {
	w.Write([]byte(err.Error()))
	return
}
