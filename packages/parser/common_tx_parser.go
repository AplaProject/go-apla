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

package parser

import (
	"errors"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

// TxParser writes transactions into the queue
func (p *Parser) TxParser(hash, binaryTx []byte, myTx bool) error {
	// get parameters for "struct" transactions
	logger := p.GetLogger()
	txType, keyID := GetTxTypeAndUserID(binaryTx)

	header, err := CheckTransaction(binaryTx)
	if err != nil {
		p.processBadTransaction(hash, err.Error())
		return err
	}

	if !( /*txType > 127 ||*/ consts.IsStruct(int(txType))) {
		if header == nil {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("tx header is nil")
			return utils.ErrInfo(errors.New("header is nil"))
		}
		keyID = header.KeyID
	}

	if keyID == 0 {
		errStr := "undefined keyID"
		p.processBadTransaction(hash, errStr)
		return errors.New(errStr)
	}

	tx := &model.Transaction{}
	_, err = tx.Get(hash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting transaction by hash")
		return utils.ErrInfo(err)
	}
	counter := tx.Counter
	counter++
	_, err = model.DeleteTransactionByHash(hash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction by hash")
		return utils.ErrInfo(err)
	}

	// put with verified=1
	newTx := &model.Transaction{
		Hash:     hash,
		Data:     binaryTx,
		Type:     int8(txType),
		KeyID:    keyID,
		Counter:  counter,
		Verified: 1,
	}
	err = newTx.Create()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating new transaction")
		return utils.ErrInfo(err)
	}

	// remove transaction from the queue (with verified=0)
	err = p.DeleteQueueTx(hash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction from queue")
		return utils.ErrInfo(err)
	}

	return nil
}

func (p *Parser) processBadTransaction(hash []byte, errText string) error {
	logger := p.GetLogger()
	if len(errText) > 255 {
		errText = errText[:255]
	}
	// looks like there is not hash in queue_tx in this moment
	qtx := &model.QueueTx{}
	/*found*/ _, err := qtx.GetByHash(hash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting tx by hash from queue")
	}

	p.DeleteQueueTx(hash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction from queue")
		return utils.ErrInfo(err)
	}
	// -----
	if qtx.FromGate == 0 {
		m := &model.TransactionStatus{}
		err = m.SetError(errText, hash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("setting transaction status error")
			return utils.ErrInfo(err)
		}
	}
	return nil
}

// DeleteQueueTx deletes a transaction from the queue
func (p *Parser) DeleteQueueTx(hash []byte) error {
	logger := p.GetLogger()
	delQueueTx := &model.QueueTx{Hash: hash}
	err := delQueueTx.DeleteTx()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction from queue")
		return utils.ErrInfo(err)
	}
	// Because we process transactions with verified=0 in queue_parser_tx, after processing we need to delete them
	_, err = model.DeleteTransactionIfUnused(hash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction if unused")
		return utils.ErrInfo(err)
	}
	return nil
}

// AllTxParser parses new transactions
func (p *Parser) AllTxParser() error {
	logger := p.GetLogger()
	all, err := model.GetAllUnverifiedAndUnusedTransactions()
	for _, data := range all {
		err = p.TxParser(data.Hash, data.Data, false)
		if err != nil {
			return utils.ErrInfo(err)
		}
		logger.Debug("transaction parsed successfully")
	}
	return nil
}
