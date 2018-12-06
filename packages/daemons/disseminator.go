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
	"context"

	"github.com/AplaProject/go-apla/packages/network/tcpclient"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/nodeban"
	"github.com/AplaProject/go-apla/packages/queue"

	log "github.com/sirupsen/logrus"
)

// Disseminator is send to all nodes from nodes_connections the following data
// if we are full node(miner): sends blocks and transactions hashes
// else send the full transactions
func Disseminator(ctx context.Context, d *daemon) error {
	DBLock()
	defer DBUnlock()

	isFullNode := true
	myNodePosition, err := syspar.GetNodePositionByKeyID(conf.Config.KeyID)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Debug("finding node")
		isFullNode = false
	}

	if isFullNode {
		// send blocks and transactions hashes
		d.logger.Debug("we are full_node, sending hashes")
		return sendBlockWithTxHashes(ctx, myNodePosition, d.logger)
	}

	// we are not full node for this StateID and WalletID, so just send transactions
	d.logger.Debug("we are full_node, sending transactions")
	return sendTransactions(ctx, d.logger)
}

func sendTransactions(ctx context.Context, logger *log.Entry) error {
	// get unsent transactions
	// form packet to send
	return queue.SendTxQueue.ProcessAllItems(func(txs []*blockchain.Transaction) error {
		hosts, err := nodeban.GetNodesBanService().FilterBannedHosts(syspar.GetRemoteHosts())
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("on getting remotes hosts")
			return err
		}

		if err := tcpclient.SendTransacitionsToAll(ctx, hosts, txs); err != nil {
			log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending transactions")
			return err
		}
		return nil
	})
}

// send block and transactions hashes
func sendBlockWithTxHashes(ctx context.Context, fullNodeID int64, logger *log.Entry) error {
	return queue.SendTxQueue.ProcessAllItems(func(trs []*blockchain.Transaction) error {
		block, isEmpty, err := queue.SendBlockQueue.Dequeue()
		if err != nil {
			return err
		}
		if isEmpty {
			return nil
		}
		if len(trs) == 0 && block == nil {
			// it's nothing to send
			logger.Debug("nothing to send")
			return nil
		}

		hosts, err := nodeban.GetNodesBanService().FilterBannedHosts(syspar.GetRemoteHosts())
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("on getting remotes hosts")
			return err
		}

		if err := tcpclient.SendFullBlockToAll(ctx, hosts, block, trs, fullNodeID); err != nil {
			log.WithFields(log.Fields{"type": consts.TCPClientError, "error": err}).Warn("on sending block with hashes to all")
			return err
		}
		return nil
	})

}
