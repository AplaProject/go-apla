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
	"fmt"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/logging"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
)

// TxParser writes transactions into the queue
func (p *Parser) TxParser(hash, binaryTx []byte, myTx bool) error {
	var err error
	var fatalError string
	var header *tx.Header

	txType, walletID, citizenID := GetTxTypeAndUserID(binaryTx)
	p.BinaryData = binaryTx
	p.TxBinaryData = binaryTx
	header, err = p.ParseDataGate(false)

	if err != nil || len(fatalError) > 0 {
		p.DeleteQueueTx(hash) // удалим тр-ию из очереди
		// remove transaction from the turn
	}
	if err == nil && len(fatalError) > 0 {
		err = errors.New(fatalError)
	}

	if err != nil {
		log.Error("err: %v", err)
		errText := fmt.Sprintf("%s", err)
		if len(errText) > 255 {
			errText = errText[:255]
		}
		qtx := &model.QueueTx{}
		err = qtx.GetByHash(hash)
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("fromGate %d", qtx.FromGate)
		if qtx.FromGate == 0 {
			m := &model.TransactionStatus{}
			err = m.SetError(errText, hash)
			if err != nil {
				return utils.ErrInfo(err)
			}
		}
	} else {
		if !( /*txType > 127 ||*/ consts.IsStruct(int(txType))) {
			if header == nil {
				return utils.ErrInfo(errors.New("header is nil"))
			}
			walletID = header.StateID
			citizenID = header.UserID
		}

		log.Debug("SELECT counter FROM transactions WHERE hex(hash) = ?", string(hash))
		logging.WriteSelectiveLog("SELECT counter FROM transactions WHERE hex(hash) = " + string(hash))
		tx := &model.Transaction{}
		err := tx.Get(hash)
		if err != nil {
			logging.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		counter := tx.Counter
		counter++
		logging.WriteSelectiveLog("DELETE FROM transactions WHERE hex(hash) = " + string(hash))
		_, err = model.DeleteTransactionByHash(hash)
		if err != nil {
			logging.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}

		log.Debug("INSERT INTO transactions (hash, data, for_self_use, type, wallet_id, citizen_id, third_var, counter) VALUES (%s, %s, %v, %v, %v, %v, %v, %v)", hash, converter.BinToHex(binaryTx), 0, int8(txType), walletID, citizenID, 0, counter)
		logging.WriteSelectiveLog("INSERT INTO transactions (hash, data, for_self_use, type, wallet_id, citizen_id, third_var, counter) VALUES ([hex], [hex], ?, ?, ?, ?, ?, ?)")
		// вставляем с verified=1
		// put with verified=1
		newTx := &model.Transaction{
			Hash:       hash,
			Data:       converter.BinToHex(binaryTx),
			ForSelfUse: 0,
			Type:       int8(txType),
			WalletID:   walletID,
			CitizenID:  citizenID,
			ThirdVar:   0,
			Counter:    counter,
		}
		err = newTx.Create()
		if err != nil {
			logging.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		logging.WriteSelectiveLog("result insert")
		log.Debug("INSERT INTO transactions - OK")
		// удалим тр-ию из очереди (с verified=0)
		// remove transaction from the turn (with verified=0)
		err = p.DeleteQueueTx(hash)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}

// DeleteQueueTx deletes a transaction from the queue
func (p *Parser) DeleteQueueTx(hashHex []byte) error {
	log.Debug("DELETE FROM queue_tx WHERE hex(hash) = %s", hashHex)
	delQueueTx := &model.QueueTx{Hash: hashHex}
	err := delQueueTx.DeleteTx()
	if err != nil {
		return utils.ErrInfo(err)
	}
	// т.к. мы обрабатываем в queue_parser_tx тр-ии с verified=0, то после их обработки их нужно удалять.
	// Because we process transactions with verified=0 in queue_parser_tx, after processing we need to delete them
	logging.WriteSelectiveLog("DELETE FROM transactions WHERE hex(hash) = " + string(hashHex) + " AND verified=0 AND used = 0")
	_, err = model.DeleteTransactionIfUnused(hashHex)
	if err != nil {
		logging.WriteSelectiveLog(err)
		return utils.ErrInfo(err)
	}
	return nil
}

// AllTxParser parses new transactions
func (p *Parser) AllTxParser() error {
	all, err := model.GetAllUnverifiedAndUnusedTransactions()
	for _, data := range all {
		log.Debug("hash: %x", data.Hash)
		err = p.TxParser(data.Hash, data.Data, false)
		if err != nil {
			itx := &model.IncorrectTx{
				Time: time.Now().Unix(),
				Hash: converter.BinToHex(data.Hash),
				Err:  fmt.Sprintf("%s", err),
			}
			err0 := itx.Create()
			if err0 != nil {
				log.Error("%v", utils.ErrInfo(err0))
			}
			return utils.ErrInfo(err)
		}
	}
	return nil
}
