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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"

	"bytes"
	"context"
	"io"
	"sync"

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
)

const (
	FULL_REQUEST         = 1
	TRANSACTIONS_REQUEST = 2
)

// send to all nodes from nodes_connections the following data
// if we are full node(miner): sends blocks and transactions hashes
// else send the full transactions
func Disseminator(d *daemon, ctx context.Context) error {
	config := &model.Config{}
	err := config.GetConfig()
	if err != nil {
		log.Errorf("can't get config: %s", err)
		return err
	}

	systemState := &model.SystemRecognizedState{}
	delegated, err := systemState.IsDelegated(config.StateID)
	if err != nil {
		log.Errorf("can't get delegated status: %s", err)
		return err
	}

	node := &model.FullNode{}
	err = node.FindNode(config.StateID, config.DltWalletID, config.StateID, config.DltWalletID)
	if err != nil {
		log.Errorf("can't get full_node: %s", err)
		return err
	}
	fullNodeID := node.ID

	// find out who we are, fullnode or not
	isFullNode := func() bool {
		if config.StateID > 0 && delegated {
			// we are state and we have delegated some work to another node
			return false
		}
		if fullNodeID == 0 {
			return false
		}
		return true
	}()

	if isFullNode {
		// send blocks and transactions hashes
		log.Debugf("we are full_node")
		return sendHashes(fullNodeID)
	} else {
		// we are not full node for this StateID and WalletID, so just send transactions
		log.Debugf("we are not full_node")
		return sendTransactions()
	}
}

func sendTransactions() error {
	// get unsent transactions
	trs, err := model.GetAllUnsentTransactions()

	if err != nil {
		return err
	}

	if trs == nil {
		return nil
	}

	// form packet to send
	var buf bytes.Buffer
	for _, tr := range *trs {
		buf.Write(MarshallTr(tr))
	}

	if buf.Len() > 0 {
		err := sendPacketToAll(TRANSACTIONS_REQUEST, buf.Bytes(), nil)
		if err != nil {
			return err
		}
	}

	// set all transactions as sent
	for _, tr := range *trs {
		_, err := model.MarkTransactionSent(tr.Hash)
		if err != nil {
			log.Errorf("failed to set transaction as sent: %s", err)
		}
	}

	return nil
}

// send block and transactions hashes
func sendHashes(fullNodeID int32) error {
	block, err := model.BlockGetUnsent()
	if err != nil {
		return err
	}

	trs, err := model.GetAllUnsentTransactions()
	if err != nil {
		return err
	}

	if (trs == nil || len(*trs) == 0) && block == nil {
		// it's nothing to send
		log.Debugf("it's nothing to send")
		return nil
	}

	buf := prepareHashReq(block, trs, fullNodeID)
	if buf != nil || len(buf) > 0 {
		err := sendPacketToAll(FULL_REQUEST, buf, sendHashesResp)
		if err != nil {
			return err
		}
	}

	// mark all transactions and block as sent
	if block != nil {
		err = block.MarkSent()
		if err != nil {
			return err
		}
	}

	if trs != nil {
		for _, tr := range *trs {
			_, err := model.MarkTransactionSent(tr.Hash)
			if err != nil {
				log.Errorf("error set transaction %+v as sent: %s", tr, err)
			}
		}
	}

	return nil
}

func sendHashesResp(resp []byte, w io.Writer) error {
	var buf bytes.Buffer
	for len(resp) > 16 {
		// Parse the list of requested transactions
		txHash := converter.BytesShift(&resp, 16)
		tr := &model.Transaction{}
		err := tr.Read(txHash)
		if err != nil {
			return err
		}
		if len(tr.Data) > 0 {
			buf.Write(converter.EncodeLengthPlusData(tr.Data))
		}
	}
	// write out the requested transactions
	_, err := w.Write(converter.DecToBin(buf.Len(), 4))
	if err != nil {
		return err
	}
	_, err = w.Write(buf.Bytes())
	return err
}

func prepareHashReq(block *model.InfoBlock, trs *[]model.Transaction, nodeID int32) []byte {
	var noBlockFlag byte
	if block != nil {
		noBlockFlag = 1
	}

	var buf bytes.Buffer
	buf.Write(converter.DecToBin(nodeID, 2))
	buf.WriteByte(noBlockFlag)
	if block != nil {
		buf.Write(MarshallBlock(block))
	}
	if trs != nil {
		for _, tr := range *trs {
			buf.Write(MarshallTrHash(tr))
		}
	}

	return buf.Bytes()
}

func MarshallTr(tr model.Transaction) []byte {
	return tr.Data

}

func MarshallBlock(block *model.InfoBlock) []byte {
	if block != nil {
		toBeSent := converter.DecToBin(block.BlockID, 3)
		return append(toBeSent, block.Hash...)
	}
	return []byte{}
}

func MarshallTrHash(tr model.Transaction) []byte {
	return tr.Hash
}

func sendPacketToAll(reqType int, buf []byte, respHand func(resp []byte, w io.Writer) error) error {
	hosts, err := model.GetFullNodesHosts()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			err := sendDRequest(h, reqType, buf, respHand)
			if err != nil {
				log.Infof("failed to send transaction to %s (%s)", h, err)
			}
			wg.Done()
		}(GetHostPort(host))
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

func sendDRequest(host string, reqType int, buf []byte, respHandler func([]byte, io.Writer) error) error {
	conn, err := utils.TCPConn(host)
	if err != nil {
		return err
	}
	defer conn.Close()

	// type
	_, err = conn.Write(converter.DecToBin(reqType, 2))
	if err != nil {
		return err
	}

	// data size
	size := converter.DecToBin(len(buf), 4)
	_, err = conn.Write(size)
	if err != nil {
		return err
	}

	// data
	_, err = conn.Write(buf)
	if err != nil {
		return err
	}

	// if response handler exist, read the answer and call handler
	if respHandler != nil {
		buf := make([]byte, 4)
		// read data size
		_, err = io.ReadFull(conn, buf)

		respSize := converter.BinToDec(buf)
		if respSize > syspar.GetMaxTxSize() {
			return nil
		}
		// read the data
		resp := make([]byte, respSize)
		_, err = io.ReadFull(conn, resp)
		if err != nil {
			return err
		}
		err = respHandler(resp, conn)
		if err != nil {
			return err
		}
	}
	return nil
}
