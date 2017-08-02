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
	"net"
	"time"

	"context"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/tcpserver"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

var tick int

// Confirmations gets and checks blocks from nodes
// Getting amount of nodes, which has the same hash as we do
func Confirmations(d *daemon, ctx context.Context) error {

	// the first 2 minutes we sleep for 10 sec for blocks to be collected
	tick++

	d.sleepTime = 1
	if tick < 12 {
		d.sleepTime = 10
	}

	var startBlockID int64

	// check last blocks, but not more than 5
	confirmations := &model.Confirmation{}
	err := confirmations.GetGoodBlock(consts.MIN_CONFIRMED_NODES)
	if err != nil {
		logger.Error("%v", err)
		return err
	}

	ConfirmedBlockID := confirmations.BlockID
	infoBlock := &model.InfoBlock{}
	err = infoBlock.GetInfoBlock()
	if err != nil {
		logger.Error("%v", err)
	}
	LastBlockID := infoBlock.BlockID
	if LastBlockID-ConfirmedBlockID > 5 {
		startBlockID = ConfirmedBlockID + 1
		d.sleepTime = 10
		tick = 0 // reset the tick
	}
	if startBlockID == 0 {
		startBlockID = LastBlockID - 1
	}
	logger.Debug("startBlockID: %d / LastBlockID: %d", startBlockID, LastBlockID)

	for blockID := LastBlockID; blockID > startBlockID; blockID-- {

		if ctx.Err() != nil {
			return ctx.Err()
		}

		logger.Debug("blockID: %d", blockID)

		block := model.Block{}
		err := block.GetBlock(blockID)
		if err != nil {
			logger.Error("%v", err)
			return err
		}
		hash := string(block.Hash)
		logger.Info("hash: %x", hash)
		if len(hash) == 0 {
			logger.Debug("len(hash) == 0")
			continue
		}

		var hosts []string
		if d.ConfigIni["test_mode"] == "1" {
			hosts = []string{"localhost"}
		} else {
			hosts, err = model.GetFullNodesHosts()
			if err != nil {
				logger.Error("%v", err)
				return err
			}
		}

		ch := make(chan string)
		for i := 0; i < len(hosts); i++ {
			// TODO: ports should be in the table hosts
			host := hosts[i] + ":" + utils.GetTcpPort(d.ConfigIni)
			logger.Info("host %v", host)
			go func() {
				IsReachable(host, blockID, ch)
			}()
		}
		var answer string
		var st0, st1 int64
		for i := 0; i < len(hosts); i++ {
			answer = <-ch
			logger.Info("answer == hash (%x = %x)", answer, hash)
			logger.Info("answer == hash (%s = %s)", answer, hash)
			if answer == hash {
				st1++
			} else {
				st0++
			}
			logger.Info("st0 %v  st1 %v", st0, st1)
		}
		confirmation := &model.Confirmation{}
		err = confirmation.GetConfirmation(blockID)
		if err == nil {
			logger.Debug("UPDATE confirmations SET good = %v, bad = %v, time = %v WHERE block_id = %v", st1, st0, time.Now().Unix(), blockID)
			confirmation.Good = int32(st1)
			confirmation.Bad = int32(st0)
			confirmation.Time = int32(time.Now().Unix())
			err = confirmation.Save()
			if err != nil {
				logger.Error("%v", err)
			}
		} else {
			confirmation.Good = int32(st1)
			confirmation.Bad = int32(st0)
			confirmation.Time = int32(time.Now().Unix())
			logger.Debug("INSERT INTO confirmations ( block_id, good, bad, time ) VALUES ( %v, %v, %v, %v )", blockID, st1, st0, time.Now().Unix())
			err = confirmation.Save()
			if err != nil {
				logger.Error("%v", err)
			}
		}
		logger.Debug("blockID > startBlockID && st1 >= consts.MIN_CONFIRMED_NODES %d>%d && %d>=%d\n", blockID, startBlockID, st1, consts.MIN_CONFIRMED_NODES)
		if blockID > startBlockID && st1 >= consts.MIN_CONFIRMED_NODES {
			break
		}
	}

	return nil

}

func checkConf(host string, blockID int64) string {
	logger.Debug("host: %v", host)
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		logger.Debug("%v", utils.ErrInfo(err))
		return "0"
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(consts.READ_TIMEOUT * time.Second))
	conn.SetWriteDeadline(time.Now().Add(consts.WRITE_TIMEOUT * time.Second))

	type confRequest struct {
		Type    uint16
		BlockID uint32
	}
	err = tcpserver.SendRequest(&confRequest{Type: 4, BlockID: uint32(blockID)}, conn)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return "0"
	}

	resp := &tcpserver.ConfirmResponse{}
	err = tcpserver.ReadRequest(resp, conn)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return "0"
	}
	return string(resp.Hash)
}

// IsReachable checks if there is blockID on the host
func IsReachable(host string, blockID int64, ch0 chan string) {
	logger.Info("IsReachable %v", host)
	ch := make(chan string, 1)
	go func() {
		ch <- checkConf(host, blockID)
	}()
	select {
	case reachable := <-ch:
		ch0 <- reachable
	case <-time.After(consts.WAIT_CONFIRMED_NODES * time.Second):
		ch0 <- "0"
	}
}
