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

package tcpserver

import (
	"bytes"
	"errors"
	"io"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

// get the list of transactions which belong to the sender from 'disseminator' daemon
// do not load the blocks here because here could be the chain of blocks that are loaded for a long time
// download the transactions here, because they are small and definitely will be downloaded in 60 sec
func Type1(r *DisRequest, rw io.ReadWriter) error {

	buf := bytes.NewBuffer(r.Data)

	/*
	 *  data structure
	 *  type - 1 byte. 0 - block, 1 - list of transactions
	 *  {if type==1}:
	 *  <any number of the next sets>
	 *   tx_hash - 16 bytes
	 * </>
	 * {if type==0}:
	 *  block_id - 3 bytes
	 *  hash - 32 bytes
	 * <any number of the next sets>
	 *   tx_hash - 16 bytes
	 * </>
	 * */

	// full_node_id of the sender to know where to take a data when it will be downloaded by another daemon
	fullNodeID := converter.BinToDec(buf.Next(8))
	log.Debug("fullNodeID", fullNodeID)

	// get data type (0 - block and transactions, 1 - only transactions)
	newDataType := converter.BinToDec(buf.Next(1))

	log.Debug("newDataType", newDataType)
	if newDataType == 0 {
		err := processBlock(buf, fullNodeID)
		if err != nil {
			return err
		}
	}

	// get unknown transactions from received packet
	needTx, err := getUnknownTransactions(buf)
	if err != nil {
		return err
	}

	// send the list of transactions which we want to get
	err = SendRequest(&DisHashResponse{Data: needTx}, rw)
	if err != nil {
		return err
	}

	if len(needTx) == 0 {
		return nil
	}

	// get this new transactions
	trs := &DisRequest{}
	err = ReadRequest(trs, rw)
	if err != nil {
		return err
	}

	// and save them
	return saveNewTransactions(trs)
}

func processBlock(buf *bytes.Buffer, fullNodeID int64) error {
	infoBlock := &model.InfoBlock{}
	found, err := infoBlock.Get()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting cur block ID")
		return utils.ErrInfo(err)
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("cant find info block")
		return errors.New("can't find info block")
	}

	// get block ID
	newBlockID := converter.BinToDec(buf.Next(3))
	log.WithFields(log.Fields{"new_block_id": newBlockID}).Debug("Generated new block id")

	// get block hash
	blockHash := buf.Next(32)
	log.Debug("blockHash %x", blockHash)

	// we accept only new blocks
	if newBlockID >= infoBlock.BlockID {
		queueBlock := &model.QueueBlock{Hash: blockHash, FullNodeID: fullNodeID, BlockID: newBlockID}
		err = queueBlock.Create()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Creating QueueBlock")
			return utils.ErrInfo(err)
		}
	}

	return nil
}

func getUnknownTransactions(buf *bytes.Buffer) ([]byte, error) {

	var needTx []byte
	for buf.Len() > 0 {
		newDataTxHash := buf.Next(16)
		if len(newDataTxHash) == 0 {
			log.WithFields(log.Fields{"len": len(newDataTxHash), "type": consts.ProtocolError}).Error("wrong transactions hash size")
			return nil, errors.New("wrong transactions hash size")
		}

		// check if we have such a transaction
		// check log_transaction
		exists, err := model.GetLogTransactionsCount(newDataTxHash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err, "txHash": newDataTxHash}).Error("Getting log tx count")
			return nil, utils.ErrInfo(err)
		}
		if exists > 0 {
			log.WithFields(log.Fields{"txHash": newDataTxHash, "type": consts.DuplicateObject}).Warning("tx with this hash already exists in log_tx")
			continue
		}

		exists, err = model.GetTransactionsCount(newDataTxHash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err, "txHash": newDataTxHash}).Error("Getting tx count")
			return nil, utils.ErrInfo(err)
		}
		if exists > 0 {
			log.WithFields(log.Fields{"txHash": newDataTxHash, "type": consts.DuplicateObject}).Warning("tx with this hash already exists in tx")
			continue
		}

		// check transaction queue
		exists, err = model.GetQueuedTransactionsCount(newDataTxHash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting queue_tx count")
			return nil, utils.ErrInfo(err)
		}
		if exists > 0 {
			log.WithFields(log.Fields{"txHash": newDataTxHash, "type": consts.DuplicateObject}).Warning("tx with this hash already exists in queue_tx")
			continue
		}
		needTx = append(needTx, newDataTxHash...)
	}

	return needTx, nil
}

func saveNewTransactions(r *DisRequest) error {

	binaryTxs := r.Data
	log.WithFields(log.Fields{"binaryTxs": binaryTxs}).Debug("trying to save binary txs")

	for len(binaryTxs) > 0 {
		txSize, err := converter.DecodeLength(&binaryTxs)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ProtocolError, "err": err}).Error("decoding binary txs length")
			return err
		}
		if int64(len(binaryTxs)) < txSize {
			log.WithFields(log.Fields{"type": consts.ProtocolError, "size": txSize, "len": len(binaryTxs)}).Error("incorrect binary txs len")
			return utils.ErrInfo(errors.New("bad transactions packet"))
		}

		txBinData := converter.BytesShift(&binaryTxs, txSize)
		if len(txBinData) == 0 {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("binaryTxs is empty")
			return utils.ErrInfo(errors.New("len(txBinData) == 0"))
		}

		if int64(len(txBinData)) > consts.MAX_TX_SIZE {
			log.WithFields(log.Fields{"type": consts.ParameterExceeded, "len": len(txBinData), "size": consts.MAX_TX_SIZE}).Error("len of tx data exceeds max size")
			return utils.ErrInfo("len(txBinData) > max_tx_size")
		}

		hash, err := crypto.Hash(txBinData)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.CryptoError, "error": err, "value": txBinData}).Fatal("cannot hash bindata")
		}

		queueTx := &model.QueueTx{Hash: hash, Data: txBinData, FromGate: 1}
		err = queueTx.Create()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("error creating QueueTx")
			return err
		}
	}
	return nil
}
