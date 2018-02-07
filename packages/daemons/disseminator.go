//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package daemons

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/GenesisCommunity/go-genesis/packages/conf"
	"github.com/GenesisCommunity/go-genesis/packages/config/syspar"
	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/model"
	"github.com/GenesisCommunity/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
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
	trs, err := model.GetAllUnsentTransactions()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all unsent transactions")
		return err
	}
	if trs == nil {
		logger.Info("transactions not found")
		return nil
	}

	// form packet to send
	var buf bytes.Buffer
	for _, tr := range *trs {
		buf.Write(MarshallTr(tr))
	}

	if buf.Len() > 0 {
		err := sendPacketToAll(I_AM_NOT_FULL_NODE, buf.Bytes(), nil, logger)
		if err != nil {
			return err
		}
	}

	// set all transactions as sent
	for _, tr := range *trs {
		_, err := model.MarkTransactionSent(tr.Hash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("marking transaction sent")
		}
	}

	return nil
}

// send block and transactions hashes
func sendHashes(fullNodeID int64, logger *log.Entry) error {
	block, err := model.BlockGetUnsent()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting unsent blocks")
		return err
	}

	trs, err := model.GetAllUnsentTransactions()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting unsent transactions")
		return err
	}

	if (trs == nil || len(*trs) == 0) && block == nil {
		// it's nothing to send
		logger.Debug("nothing to send")
		return nil
	}

	buf := prepareHashReq(block, trs, fullNodeID)
	if buf != nil || len(buf) > 0 {
		err := sendPacketToAll(I_AM_FULL_NODE, buf, sendHashesResp, logger)
		if err != nil {
			return err
		}
	}

	// mark all transactions and block as sent
	if block != nil {
		err = block.MarkSent()
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("marking block sent")
			return err
		}
	}

	if trs != nil {
		for _, tr := range *trs {
			_, err := model.MarkTransactionSent(tr.Hash)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("marking transaction sent")
			}
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

func prepareHashReq(block *model.InfoBlock, trs *[]model.Transaction, nodeID int64) []byte {
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
		for _, tr := range *trs {
			buf.Write(MarshallTrHash(tr))
		}
	}

	return buf.Bytes()
}

// MarshallTr returns transaction data
func MarshallTr(tr model.Transaction) []byte {
	return tr.Data
}

// MarshallBlock returns block as []byte
func MarshallBlock(block *model.InfoBlock) []byte {
	if block != nil {
		toBeSent := converter.DecToBin(block.BlockID, 3)
		return append(toBeSent, block.Hash...)
	}
	return []byte{}
}

// MarshallTrHash returns transaction hash
func MarshallTrHash(tr model.Transaction) []byte {
	return tr.Hash
}

func sendPacketToAll(reqType int, buf []byte, respHand func(resp []byte, w io.Writer, logger *log.Entry) error, logger *log.Entry) error {

	hosts := syspar.GetRemoteHosts()
	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			sendDRequest(h, reqType, buf, respHand, logger)
			wg.Done()
		}(getHostPort(host))
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
