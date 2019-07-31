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

package rollback

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/AplaProject/go-apla/packages/block"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

var (
	ErrLastBlock = errors.New("Block is not the last")
)

// BlockRollback is blocking rollback
func RollbackBlock(data []byte) error {
	bl, err := block.UnmarshallBlock(bytes.NewBuffer(data), true)
	if err != nil {
		return err
	}

	b := &model.Block{}
	if _, err = b.GetMaxBlock(); err != nil {
		return err
	}

	if b.ID != bl.Header.BlockID {
		return ErrLastBlock
	}

	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting transaction")
		return err
	}

	err = rollbackBlock(dbTransaction, bl)
	if err != nil {
		dbTransaction.Rollback()
		return err
	}

	if err = b.DeleteById(dbTransaction, bl.Header.BlockID); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting block by id")
		dbTransaction.Rollback()
		return err
	}

	b = &model.Block{}
	if _, err = b.Get(bl.Header.BlockID - 1); err != nil {
		dbTransaction.Rollback()
		return err
	}

	bl, err = block.UnmarshallBlock(bytes.NewBuffer(b.Data), false)
	if err != nil {
		dbTransaction.Rollback()
		return err
	}

	ib := &model.InfoBlock{
		Hash:           b.Hash,
		RollbacksHash:  b.RollbacksHash,
		BlockID:        b.ID,
		NodePosition:   strconv.Itoa(int(b.NodePosition)),
		KeyID:          b.KeyID,
		Time:           b.Time,
		CurrentVersion: strconv.Itoa(bl.Header.Version),
	}
	err = ib.Update(dbTransaction)
	if err != nil {
		dbTransaction.Rollback()
		return err
	}

	return dbTransaction.Commit()
}

func rollbackBlock(dbTransaction *model.DbTransaction, block *block.Block) error {
	// rollback transactions in reverse order
	logger := block.GetLogger()
	for i := len(block.Transactions) - 1; i >= 0; i-- {
		t := block.Transactions[i]
		t.DbTransaction = dbTransaction

		_, err := model.MarkTransactionUnusedAndUnverified(dbTransaction, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting transaction")
			return err
		}
		_, err = model.DeleteLogTransactionsByHash(dbTransaction, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting log transactions by hash")
			return err
		}

		ts := &model.TransactionStatus{}
		err = ts.UpdateBlockID(dbTransaction, 0, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating block id in transaction status")
			return err
		}

		_, err = model.DeleteQueueTxByHash(dbTransaction, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transacion from queue by hash")
			return err
		}

		if t.TxContract != nil {
			if err = rollbackTransaction(t.TxHash, t.DbTransaction, logger); err != nil {
				return err
			}
		} else {
			MethodName := consts.TxTypes[t.TxType]
			txParser, err := transaction.GetTransaction(t, MethodName)
			if err != nil {
				return utils.ErrInfo(err)
			}
			result := txParser.Init()
			if _, ok := result.(error); ok {
				return utils.ErrInfo(result.(error))
			}
			result = txParser.Rollback()
			if _, ok := result.(error); ok {
				return utils.ErrInfo(result.(error))
			}
		}
	}

	return nil
}
