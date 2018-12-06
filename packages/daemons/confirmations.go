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
	"time"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/network/tcpclient"
	"github.com/AplaProject/go-apla/packages/nodeban"

	log "github.com/sirupsen/logrus"
)

var tick int

// Confirmations gets and checks blocks from nodes
// Getting amount of nodes, which has the same hash as we do
func Confirmations(ctx context.Context, d *daemon) error {

	// the first 2 minutes we sleep for 10 sec for blocks to be collected
	tick++

	d.sleepTime = 1 * time.Second
	if tick < 12 {
		d.sleepTime = 10 * time.Second
	}

	blocks, err := blockchain.GetUnconfirmedBlocks(nil, consts.MIN_CONFIRMED_NODES)
	if err != nil {
		return err
	}
	if len(blocks) == 0 {
		return nil
	}

	return confirmationsBlocks(ctx, d, blocks)
}

func confirmationsBlocks(ctx context.Context, d *daemon, blocks []*blockchain.BlockWithHash) error {
	for _, block := range blocks {
		if err := ctx.Err(); err != nil {
			d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": err}).Error("error in context")
			return err
		}

		hashStr := string(converter.BinToHex(block.Hash))
		d.logger.WithFields(log.Fields{"hash": hashStr}).Debug("checking hash")

		hosts, err := nodeban.GetNodesBanService().FilterBannedHosts(syspar.GetRemoteHosts())
		if err != nil {
			return err
		}

		ch := make(chan string)
		for i := 0; i < len(hosts); i++ {
			host, err := tcpclient.NormalizeHostAddress(hosts[i], consts.DEFAULT_TCP_PORT)
			if err != nil {
				d.logger.WithFields(log.Fields{"host": host[i], "type": consts.ParseError, "error": err}).Error("wrong host address")
				continue
			}

			d.logger.WithFields(log.Fields{"host": host, "block_hash": block.Hash}).Debug("checking block id confirmed at node")
			go func() {
				IsReachable(host, block.Hash, ch, d.logger)
			}()
		}
		var answer string
		var st0, st1 int
		for i := 0; i < len(hosts); i++ {
			answer = <-ch
			if answer == hashStr {
				st1++
			} else {
				st0++
			}
		}
		confirmation := &blockchain.Confirmation{}
		confirmation.BlockID = block.Block.Header.BlockID
		confirmation.Good = st1
		confirmation.Bad = st0
		confirmation.Time = time.Now().Unix()
		if err = confirmation.Insert(nil, block.Hash); err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("saving confirmation")
			return err
		}
	}

	return nil
}

// IsReachable checks if there is blockID on the host
func IsReachable(host string, blockHash []byte, ch0 chan string, logger *log.Entry) {
	ch := make(chan string, 1)
	go func() {
		ch <- tcpclient.CheckConfirmation(host, blockHash, logger)
	}()
	select {
	case reachable := <-ch:
		ch0 <- reachable
	case <-time.After(consts.WAIT_CONFIRMED_NODES * time.Second):
		ch0 <- "0"
	}
}
