// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package rollback

import (
	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

// ToBlockID rollbacks blocks till blockID
func ToBlockID(blockHash []byte, dbTransaction *model.DbTransaction, ldbTx *leveldb.Transaction, logger *log.Entry) error {
	blocks, err := blockchain.DeleteBlocksFrom(nil, blockHash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting blocks")
		return err
	}
	if len(blocks) == 0 {
		return nil
	}
	for _, block := range blocks {
		// roll back our blocks to the block blockID
		err = RollbackBlock(block.Block, block.Hash)
		if err != nil {
			return err
		}
	}

	return nil
}
