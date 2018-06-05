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
	"bytes"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/smart"

	log "github.com/sirupsen/logrus"
)

// BlockRollback is blocking rollback
func RollbackBlock(data []byte, deleteBlock bool) error {
	buf := bytes.NewBuffer(data)
	if buf.Len() == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty buffer")
		return fmt.Errorf("empty buffer")
	}

	block, err := parser.ParseBlock(buf, false)
	if err != nil {
		return err
	}

	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting transaction")
		return err
	}

	err = rollbackBlock(dbTransaction, block)

	if err != nil {
		dbTransaction.Rollback()
		return err
	}

	if deleteBlock {
		b := &model.Block{}
		err = b.DeleteById(dbTransaction, block.Header.BlockID)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting block by id")
			dbTransaction.Rollback()
			return err
		}
	}

	err = dbTransaction.Commit()
	return err
}

func rollbackBlock(transaction *model.DbTransaction, block *parser.Block) error {
	// rollback transactions in reverse order
	logger := block.GetLogger()
	for i := len(block.Parsers) - 1; i >= 0; i-- {
		p := block.Parsers[i]
		p.DbTransaction = transaction

		_, err := model.MarkTransactionUnusedAndUnverified(transaction, p.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting transaction")
			return err
		}
		_, err = model.DeleteLogTransactionsByHash(transaction, p.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting log transactions by hash")
			return err
		}

		ts := &model.TransactionStatus{}
		err = ts.UpdateBlockID(transaction, 0, p.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating block id in transaction status")
			return err
		}

		_, err = model.DeleteQueueTxByHash(transaction, p.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transacion from queue by hash")
			return err
		}

		if p.TxContract != nil {
			if _, err := p.CallContract(smart.CallInit | smart.CallRollback); err != nil {
				return err
			}
			if err = rollbackTransaction(p.TxHash, p.DbTransaction, logger); err != nil {
				return err
			}
		} else {
			MethodName := consts.TxTypes[int(p.TxType)]
			txParser, err := parser.GetParser(p, MethodName)
			if err != nil {
				return p.ErrInfo(err)
			}
			result := txParser.Init()
			if _, ok := result.(error); ok {
				return p.ErrInfo(result.(error))
			}
			result = txParser.Rollback()
			if _, ok := result.(error); ok {
				return p.ErrInfo(result.(error))
			}
		}
	}

	return nil
}
