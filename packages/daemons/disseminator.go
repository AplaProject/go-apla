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
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"

	"bytes"
	"context"
	"io"
	"sync"

	"github.com/AplaProject/go-apla/packages/config/syspar"
)

const (
	I_AM_FULL_NODE = 1
	I_AM_NOT_FULL_NODE = 2
)

// send to all nodes from nodes_connections the following data
// if we are full node(miner): sends blocks and transactions hashes
// else send the full transactions
func Disseminator(d *daemon, ctx context.Context) error {
	config := &model.Config{}
	_, err := config.Get()
	if err != nil {
		log.Errorf("can't get config: %s", err)
		return err
	}

	isFullNode := true
	myNodePosition, err := syspar.GetNodePositionByKeyID(config.KeyID)
	if err != nil  {
		log.Error("%v", err)
		isFullNode = false
	}

	if isFullNode {
		// send blocks and transactions hashes
		log.Debugf("we are full_node")
		return sendHashes(myNodePosition)
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
		err := sendPacketToAll(I_AM_NOT_FULL_NODE, buf.Bytes(), nil)
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
func sendHashes(fullNodeID int64) error {
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
			err := sendPacketToAll(I_AM_FULL_NODE, buf, sendHashesResp)
			if err != nil {
				return err
			}
	}

	// mark all transactions and block as sent
	if block != nil {
		//err = block.MarkSent()
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
		_, err := tr.Read(txHash)
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

func prepareHashReq(block *model.InfoBlock, trs *[]model.Transaction, nodeID int64) []byte {
	var noBlockFlag byte
	if block == nil {
		noBlockFlag = 1
	}

	var buf bytes.Buffer
	buf.Write(converter.DecToBin(nodeID, 8))
	buf.WriteByte(noBlockFlag)
	if noBlockFlag==0 {
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
	hosts := syspar.GetHosts()
	log.Debug("sendPacketToAll",hosts )
	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			err := sendDRequest(h, reqType, buf, respHand)
			if err != nil {
				log.Infof("failed to send transaction to %s (%s)", h, err)
			}
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
	log.Debug("reqType", reqType)

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
