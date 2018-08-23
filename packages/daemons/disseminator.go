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

package daemons

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/queue"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

const (
	// I_AM_FULL_NODE is full node flag
	I_AM_FULL_NODE = 1
	// I_AM_NOT_FULL_NODE is not full node flag
	I_AM_NOT_FULL_NODE = 2
)

// Disseminator is send to all nodes from nodes_connections the following data
// if we are full node(miner): sends blocks and transactions hashes
// else send the full transactions
func Disseminator(ctx context.Context, d *daemon) error {
	isFullNode := true
	myNodePosition, err := syspar.GetNodePositionByKeyID(conf.Config.KeyID)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Debug("finding node")
		isFullNode = false
	}

	if isFullNode {
		// send blocks and transactions hashes
		d.logger.Debug("we are full_node, sending hashes")
		return sendHashes(myNodePosition, d.logger)
	}

	// we are not full node for this StateID and WalletID, so just send transactions
	d.logger.Debug("we are full_node, sending transactions")
	return sendTransactions(d.logger)
}

func sendTransactions(logger *log.Entry) error {
	// get unsent transactions
	// form packet to send
	var buf bytes.Buffer
	for queue.SendTxQueue.Length() > 0 {
		item, err := queue.SendTxQueue.Dequeue()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("peeking item from sendTx queue")
			return err
		}
		buf.Write(item.Value)
	}
	if buf.Len() > 0 {
		err := sendPacketToAll(I_AM_NOT_FULL_NODE, buf.Bytes(), nil, logger)
		if err != nil {
			return err
		}
	}
	return nil
}

// send block and transactions hashes
func sendHashes(fullNodeID int64, logger *log.Entry) error {
	blockItem, err := queue.SendBlockQueue.Dequeue()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("getting unsent blocks")
		return err
	}
	block := &blockchain.Block{}
	if err := block.Unmarshal(blockItem.Value); err != nil {
		logger.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling blockchain block")
	}

	var trs []*tx.SmartContract
	for queue.SendTxQueue.Length() > 0 {
		txItem, err := queue.SendTxQueue.Dequeue()
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("getting unsent blocks")
			return err
		}
		tr := &tx.SmartContract{}
		if err := msgpack.Unmarshal(txItem.Value, tr); err != nil {
			logger.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling transaction")
			return err
		}
		trs = append(trs, tr)

	}
	if len(trs) == 0 && block == nil {
		// it's nothing to send
		logger.Debug("nothing to send")
		return nil
	}

	buf, err := prepareHashReq(block, trs, fullNodeID)
	if err != nil {
		return err
	}
	if buf != nil || len(buf) > 0 {
		err := sendPacketToAll(I_AM_FULL_NODE, buf, sendHashesResp, logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendHashesResp(resp []byte, w io.Writer, logger *log.Entry) error {
	var buf bytes.Buffer
	lr := len(resp)
	switch true {
	// We got response that mean other full node have all of transactions and we don't need to do anything
	case lr == 0:
		return nil
	// We got response that mean that node doesn't know about some transactions. We need to send them
	case lr >= consts.HashSize:
		for len(resp) >= consts.HashSize {
			// Parse the list of requested transactions
			txHash := converter.BytesShift(&resp, consts.HashSize)
			tr := &model.Transaction{}
			_, err := tr.Read(txHash)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("reading transaction by hash")
				return err
			}

			if len(tr.Data) > 0 {
				buf.Write(converter.EncodeLengthPlusData(tr.Data))
			}
		}

		// write out the requested transactions
		_, err := w.Write(converter.DecToBin(buf.Len(), 4))
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing tx size")
			return err
		}
		_, err = w.Write(buf.Bytes())
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing tx data")
			return err
		}
		return nil
	}

	return nil
}

func prepareHashReq(block *blockchain.Block, trs []*tx.SmartContract, nodeID int64) ([]byte, error) {
	var noBlockFlag byte
	if block == nil {
		noBlockFlag = 1
	}

	var buf bytes.Buffer
	buf.Write(converter.DecToBin(nodeID, 8))
	buf.WriteByte(noBlockFlag)
	if noBlockFlag == 0 {
		buf.Write(MarshallBlock(block))
	}
	if trs != nil {
		for _, tr := range trs {
			trHash, err := MarshallTrHash(tr)
			if err != nil {
				return nil, err
			}
			buf.Write(trHash)
		}
	}

	return buf.Bytes(), nil
}

// MarshallTr returns transaction data
func MarshallTr(tr *tx.SmartContract) (b []byte, err error) {
	if b, err = msgpack.Marshal(tr); err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling transaction")
		return
	}
	return
}

// MarshallBlock returns block as []byte
func MarshallBlock(block *blockchain.Block) []byte {
	if block != nil {
		toBeSent := converter.DecToBin(block.Header.BlockID, 3)
		return append(toBeSent, block.Header.Hash...)
	}
	return []byte{}
}

// MarshallTrHash returns transaction hash
func MarshallTrHash(tr *tx.SmartContract) ([]byte, error) {
	b, err := MarshallTr(tr)
	if err != nil {
		return nil, err
	}
	hash, err := crypto.DoubleHash(b)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("hashing error")
		return nil, err
	}
	return hash, nil
}

func sendPacketToAll(reqType int, buf []byte, respHand func(resp []byte, w io.Writer, logger *log.Entry) error, logger *log.Entry) error {

	hosts, err := filterBannedHosts(syspar.GetRemoteHosts())
	if err != nil {
		return err
	}
	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			sendDRequest(h, reqType, buf, respHand, logger)
			wg.Done()
		}(utils.GetHostPort(host))
	}
	wg.Wait()

	return nil
}

/*
Packet format:
type  2 bytes
len   4 bytes
data  len bytes
*/

func sendDRequest(host string, reqType int, buf []byte, respHandler func([]byte, io.Writer, *log.Entry) error, logger *log.Entry) error {
	conn, err := utils.TCPConn(host)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": host}).Debug("tcp connection to host")
		return err
	}
	defer conn.Close()

	// type
	_, err = conn.Write(converter.DecToBin(reqType, 2))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Error("writing request type to host")
		return err
	}

	// data size
	size := converter.DecToBin(len(buf), 4)
	_, err = conn.Write(size)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Error("writing data size to host")
		return err
	}

	// data
	_, err = conn.Write(buf)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Error("writing data to host")
		return err
	}

	// if response handler exist, read the answer and call handler
	if respHandler != nil {
		buf := make([]byte, 4)

		// read data size
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			if err == io.EOF {
				logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Warn("connection closed unexpectedly")
			} else {
				logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Error("reading data size")
			}
		}

		respSize := converter.BinToDec(buf)
		if respSize > syspar.GetMaxTxSize() {
			logger.WithFields(log.Fields{"size": respSize, "max_size": syspar.GetMaxTxSize(), "type": consts.ParameterExceeded}).Warning("response size is larger than max tx size")
			return nil
		}
		// read the data
		resp := make([]byte, respSize)
		_, err = io.ReadFull(conn, resp)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Error("reading data")
			return err
		}
		err = respHandler(resp, conn, logger)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Error("reading data")
			return err
		}
	}
	return nil
}
