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

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/network/tcpclient"
	"github.com/GenesisKernel/go-genesis/packages/service"

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

	var startBlockID int64

	// check last blocks, but not more than 5
	confirmations := &model.Confirmation{}
	_, err := confirmations.GetGoodBlock(consts.MIN_CONFIRMED_NODES)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting good block")
		return err
	}

	ConfirmedBlockID := confirmations.BlockID
	infoBlock := &model.InfoBlock{}
	_, err = infoBlock.Get()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting info block")
		return err
	}
	lastBlockID := infoBlock.BlockID
	if lastBlockID == 0 {
		return nil
	}

	if lastBlockID-ConfirmedBlockID > 5 {
		startBlockID = ConfirmedBlockID + 1
		d.sleepTime = 10 * time.Second
		tick = 0 // reset the tick
	}

	if startBlockID == 0 {
		startBlockID = lastBlockID
	}

	return confirmationsBlocks(ctx, d, lastBlockID, startBlockID)
}

func confirmationsBlocks(ctx context.Context, d *daemon, lastBlockID, startBlockID int64) error {
	for blockID := lastBlockID; blockID >= startBlockID; blockID-- {
		if err := ctx.Err(); err != nil {
			d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": err}).Error("error in context")
			return err
		}

		block := model.Block{}
		_, err := block.Get(blockID)
		if err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block by ID")
			return err
		}

		hashStr := string(converter.BinToHex(block.Hash))
		d.logger.WithFields(log.Fields{"hash": hashStr}).Debug("checking hash")
		if len(hashStr) == 0 {
			d.logger.WithFields(log.Fields{"hash": hashStr, "type": consts.NotFound}).Debug("hash not found")
			continue
		}

		hosts, err := service.GetNodesBanService().FilterBannedHosts(syspar.GetRemoteHosts())
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

			d.logger.WithFields(log.Fields{"host": host, "block_id": blockID}).Debug("checking block id confirmed at node")
			go func() {
				IsReachable(host, blockID, ch, d.logger)
			}()
		}
		var answer string
		var st0, st1 int64
		for i := 0; i < len(hosts); i++ {
			answer = <-ch
			if answer == hashStr {
				st1++
			} else {
				st0++
			}
		}
		confirmation := &model.Confirmation{}
		confirmation.GetConfirmation(blockID)
		confirmation.BlockID = blockID
		confirmation.Good = int32(st1)
		confirmation.Bad = int32(st0)
		confirmation.Time = int32(time.Now().Unix())
		if err = confirmation.Save(); err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("saving confirmation")
			return err
		}

		if blockID > startBlockID && st1 >= consts.MIN_CONFIRMED_NODES {
			break
		}
	}

	return nil
}

// IsReachable checks if there is blockID on the host
func IsReachable(host string, blockID int64, ch0 chan string, logger *log.Entry) {
	ch := make(chan string, 1)
	go func() {
		ch <- tcpclient.CheckConfirmation(host, blockID, logger)
	}()
	select {
	case reachable := <-ch:
		ch0 <- reachable
	case <-time.After(consts.WAIT_CONFIRMED_NODES * time.Second):
		ch0 <- "0"
	}
}
