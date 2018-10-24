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
	"github.com/GenesisKernel/go-genesis/packages/block"
	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

// BlockRollback is blocking rollback
func RollbackBlock(blockModel *blockchain.Block, hash []byte) error {
	ldbTx, err := blockchain.DB.OpenTransaction()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("starting transaction")
		return err
	}
	b, err := block.FromBlockchainBlock(blockModel, hash, ldbTx)
	if err != nil {
		return err
	}

	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting transaction")
		return err
	}

	err = rollbackBlock(dbTransaction, ldbTx, b)

	if err != nil {
		dbTransaction.Rollback()
		ldbTx.Discard()
		return err
	}

	err = dbTransaction.Commit()
	err = ldbTx.Commit()
	return err
}

func rollbackBlock(dbTransaction *model.DbTransaction, ldbTx *leveldb.Transaction, block *block.PlayableBlock) error {
	// rollback transactions in reverse order
	logger := block.GetLogger()
	for i := len(block.Transactions) - 1; i >= 0; i-- {
		t := block.Transactions[i]
		t.DbTransaction = dbTransaction
		t.LdbTx = ldbTx

		if t.TxContract != nil {
			if err := rollbackTransaction(t.TxHash, t.DbTransaction, logger); err != nil {
				return err
			}
		}
	}

	return nil
}
