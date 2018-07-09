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

	"github.com/GenesisKernel/go-genesis/packages/tcpclient"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/utils"

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

	hosts, err := service.GetNodesBanService().FilterBannedHosts(syspar.GetRemoteHosts())
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("on getting remotes hosts")
		return err
	}

	cli := tcpclient.NewClient(defaultTCPClientConfig(), logger)
	if err := cli.SendTransacitionsToAll(hosts, *trs); err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending transactions")
		return err
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

func sendPacketToAll(reqType int, buf []byte, respHand func(resp []byte, w io.Writer, logger *log.Entry) error, logger *log.Entry) error {

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

		if err = respHandler(resp, conn, logger); err != nil {
			logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Error("reading data")
			return err
		}
	}
	return nil
}
