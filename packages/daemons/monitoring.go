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

package daemons

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

// Monitoring starts monitoring
func Monitoring(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	infoBlock := &model.InfoBlock{}
	_, err := infoBlock.Get()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting info block")
		logError(w, fmt.Errorf("can't get info block: %s", err))
		return
	}
	addKey(&buf, "info_block_id", infoBlock.BlockID)
	addKey(&buf, "info_block_hash", converter.BinToHex(infoBlock.Hash))
	addKey(&buf, "info_block_time", infoBlock.Time)
	addKey(&buf, "info_block_key_id", infoBlock.KeyID)
	addKey(&buf, "info_block_ecosystem_id", infoBlock.EcosystemID)
	addKey(&buf, "info_block_node_position", infoBlock.NodePosition)
	addKey(&buf, "full_nodes_count", syspar.GetNumberOfNodes())

	block := &model.Block{}
	_, err = block.GetMaxBlock()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting max block")
		logError(w, fmt.Errorf("can't get max block: %s", err))
		return
	}
	addKey(&buf, "last_block_id", block.ID)
	addKey(&buf, "last_block_hash", converter.BinToHex(block.Hash))
	addKey(&buf, "last_block_time", block.Time)
	addKey(&buf, "last_block_wallet", block.KeyID)
	addKey(&buf, "last_block_state", block)
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

func addKey(buf *bytes.Buffer, key string, value interface{}) error {
	val, err := converter.InterfaceToStr(value)
	if err != nil {
		return err
	}
	line := fmt.Sprintf("%s\t%s\n", key, val)
	buf.Write([]byte(line))
	return nil
}

func logError(w http.ResponseWriter, err error) {
	w.Write([]byte(err.Error()))
	return
}
