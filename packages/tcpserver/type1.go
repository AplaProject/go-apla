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
	"fmt"
	"io"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// get the list of transactions which belong to the sender from 'disseminator' daemon
// do not load the blocks here because here could be the chain of blocks that are loaded for a long time
// download the transactions here, because they are small and definitely will be downloaded in 60 sec
func Type1(r *DisRequest, rw io.ReadWriter) error {
	logger.LogDebug(consts.DebugMessage, "")
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
	fullNodeID := converter.BinToDec(buf.Next(2))

	// get data type (0 - block and transactions, 1 - only transactions)
	newDataType := converter.BinToDec(buf.Next(1))

	if newDataType == 0 {
		err := processBlock(buf, fullNodeID)
		if err != nil {
			logger.LogError(consts.BlockError, err)
			return err
		}
	}

	// get unknown transactions from received packet
	needTx, err := getUnknownTransactions(buf)
	if err != nil {
		logger.LogError(consts.TransactionError, err)
		return err
	}

	// send the list of transactions which we want to get
	err = SendRequest(&DisHashResponse{Data: needTx}, rw)
	if err != nil {
		logger.LogError(consts.IOError, err)
		return err
	}

	if len(needTx) == 0 {
		return nil
	}

	// get this new transactions
	trs := &DisRequest{}
	err = ReadRequest(trs, rw)
	if err != nil {
		logger.LogError(consts.IOError, err)
		return err
	}

	// and save them
	return saveNewTransactions(trs)
}

func processBlock(buf *bytes.Buffer, fullNodeID int64) error {
	logger.LogDebug(consts.DebugMessage, "")
	blockID, err := model.GetCurBlockID()
	if err != nil {
		logger.LogError(consts.DBError, err)
		return utils.ErrInfo(err)
	}

	// get block ID
	newBlockID := converter.BinToDec(buf.Next(3))
	logger.LogDebug(consts.DebugMessage, fmt.Sprintf("newDataBlockID: %d / blockID: %d", newBlockID, blockID))

	// get block hash
	blockHash := buf.Next(32)

	// we accept only new blocks
	if newBlockID >= blockID {
		queueBlock := &model.QueueBlock{Hash: blockHash, FullNodeID: fullNodeID, BlockID: newBlockID}
		err = queueBlock.Create()
		if err != nil {
			logger.LogError(consts.DBError, err)
			return utils.ErrInfo(err)
		}
	}

	return nil
}

func getUnknownTransactions(buf *bytes.Buffer) ([]byte, error) {
	logger.LogDebug(consts.FuncStarted, "")
	var needTx []byte
	for buf.Len() > 0 {
		newDataTxHash := buf.Next(16)
		if len(newDataTxHash) == 0 {
			logger.LogError(consts.TransactionError, "wrong transactions hash size")
			return nil, errors.New("wrong transactions hash size")
		}

		// check if we have such a transaction
		// check log_transaction
		exists, err := model.GetLogTransactionsCount(newDataTxHash)
		if err != nil {
			logger.LogError(consts.DBError, err)
			return nil, utils.ErrInfo(err)
		}
		if exists > 0 {
			logger.LogDebug(consts.DebugMessage, "exists")
			continue
		}

		exists, err = model.GetTransactionsCount(newDataTxHash)
		if err != nil {
			logger.LogError(consts.DBError, err)
			return nil, utils.ErrInfo(err)
		}
		if exists > 0 {
			logger.LogDebug(consts.DebugMessage, "exists")
			continue
		}

		// check transaction queue
		exists, err = model.GetQueuedTransactionsCount(newDataTxHash)
		if err != nil {
			logger.LogError(consts.DBError, err)
			return nil, utils.ErrInfo(err)
		}
		if exists > 0 {
			logger.LogDebug(consts.DebugMessage, "exists")
			continue
		}
		needTx = append(needTx, newDataTxHash...)
	}

	return needTx, nil
}

func saveNewTransactions(r *DisRequest) error {
	logger.LogDebug(consts.DebugMessage, "")
	binaryTxs := r.Data
	logger.LogError(consts.DebugMessage, fmt.Sprintf("binaryTxs %x", binaryTxs))

	for len(binaryTxs) > 0 {
		txSize, err := converter.DecodeLength(&binaryTxs)
		if err != nil {
			logger.LogError(consts.TransactionError, err)
			return err
		}
		if int64(len(binaryTxs)) < txSize {
			logger.LogError(consts.TransactionError, "bad transactions packet")
			return utils.ErrInfo(errors.New("bad transactions packet"))
		}

		txBinData := converter.BytesShift(&binaryTxs, txSize)
		if len(txBinData) == 0 {
			logger.LogError(consts.TransactionError, "len(txBinData) == 0")
			return utils.ErrInfo(errors.New("len(txBinData) == 0"))
		}

		if int64(len(txBinData)) > consts.MAX_TX_SIZE {
			logger.LogError(consts.TransactionError, "len(txBinData) > max_tx_size")
			return utils.ErrInfo("len(txBinData) > max_tx_size")
		}

		hash, err := crypto.Hash(txBinData)
		if err != nil {
			logger.LogFatal(consts.CryptoError, err)
		}

		queueTx := &model.QueueTx{Hash: hash, Data: txBinData, FromGate: 1}
		err = queueTx.Create()
		if err != nil {
			logger.LogError(consts.DBError, err)
			return err
		}
	}
	return nil
}
