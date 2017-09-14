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

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/logging"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// TxParser writes transactions into the queue
func (p *Parser) TxParser(hash, binaryTx []byte, myTx bool) error {

	log.Debugf("transaction hex data: %x", binaryTx)
	p.BinaryData = binaryTx
	p.TxBinaryData = binaryTx

	// get parameters for "struct" transactions
	txType, walletID, citizenID := GetTxTypeAndUserID(binaryTx)

	header, err := p.ParseDataGate(false)
	if err != nil {
		log.Errorf("parse data gate error: %s", err)
		p.processBadTransaction(hash, err.Error())
		return err
	}

	if !( /*txType > 127 ||*/ consts.IsStruct(int(txType))) {
		if header == nil {
			return utils.ErrInfo(errors.New("header is nil"))
		}
		walletID = header.StateID
		citizenID = header.UserID
	}

	if walletID == 0 && citizenID == 0 {
		errStr := "undefined walletID and citizenID"
		p.processBadTransaction(hash, errStr)
		return errors.New(errStr)
	}

	logging.WriteSelectiveLog("SELECT counter FROM transactions WHERE hex(hash) = " + string(hash))
	tx := &model.Transaction{}
	err = tx.Get(hash)
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

	log.Errorf("wallet id for transaction - %d", walletID)

	log.Debug("INSERT INTO transactions (hash, data, for_self_use, type, wallet_id, citizen_id, third_var, counter) VALUES (%s, %s, %v, %v, %v, %v, %v, %v)", hash, converter.BinToHex(binaryTx), 0, int8(txType), walletID, citizenID, 0, counter)
	logging.WriteSelectiveLog("INSERT INTO transactions (hash, data, for_self_use, type, wallet_id, citizen_id, third_var, counter) VALUES ([hex], [hex], ?, ?, ?, ?, ?, ?)")
	// вставляем с verified=1
	// put with verified=1
	newTx := &model.Transaction{
		Hash:      hash,
		Data:      binaryTx,
		Type:      int8(txType),
		WalletID:  walletID,
		CitizenID: citizenID,
		Counter:   counter,
		Verified:  1,
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

	return nil
}

func (p *Parser) processBadTransaction(hash []byte, errText string) error {
	if len(errText) > 255 {
		errText = errText[:255]
	}
	qtx := &model.QueueTx{}
	found, err := qtx.GetByHash(hash)
	if !found {
		return nil
	}
	if err != nil {
		return utils.ErrInfo(err)
	}

	if qtx.FromGate == 0 {
		m := &model.TransactionStatus{}
		err = m.SetError(errText, hash)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	p.DeleteQueueTx(hash)
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
			log.Errorf("transaction parser error: %s", err)
			/*
				TODO: log bad transaction
				itx := &model.IncorrectTx{
					Time: time.Now().Unix(),
					Hash: converter.BinToHex(data.Hash),
					Err:  fmt.Sprintf("%s", err),
				}
				err0 := itx.Create()
				if err0 != nil {
					log.Error("can't insert incorrect transaction: %v", utils.ErrInfo(err0))
				}
			*/
			// TODO: return after first bad transaction ?
			return utils.ErrInfo(err)
		}
		log.Debugf("transaction parsed successfully")
	}
	return nil
}
