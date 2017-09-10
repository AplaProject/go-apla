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
	"fmt"
	"net"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/tcpserver"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

var tick int

// Confirmations gets and checks blocks from nodes
// Getting amount of nodes, which has the same hash as we do
func Confirmations(d *daemon, ctx context.Context) error {
	logger.LogDebug(consts.FuncStarted, "")
	// the first 2 minutes we sleep for 10 sec for blocks to be collected
	tick++

	d.sleepTime = 1 * time.Second
	if tick < 12 {
		d.sleepTime = 10 * time.Second
	}

	var startBlockID int64

	// check last blocks, but not more than 5
	confirmations := &model.Confirmation{}
	err := confirmations.GetGoodBlock(consts.MIN_CONFIRMED_NODES)
	if err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}

	ConfirmedBlockID := confirmations.BlockID
	infoBlock := &model.InfoBlock{}
	err = infoBlock.GetInfoBlock()
	if err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}
	LastBlockID := infoBlock.BlockID
	if LastBlockID == 0 {
		return nil
	}

	if LastBlockID-ConfirmedBlockID > 5 {
		startBlockID = ConfirmedBlockID + 1
		d.sleepTime = 10 * time.Second
		tick = 0 // reset the tick
	}
	if startBlockID == 0 {
		startBlockID = LastBlockID
	}
	logger.LogDebug(consts.DebugMessage, fmt.Sprintf("confirmation: startBlockID: %d / LastBlockID: %d", startBlockID, LastBlockID))

	for blockID := LastBlockID; blockID >= startBlockID; blockID-- {

		if ctx.Err() != nil {
			logger.LogError(consts.ContextError, err)
			return ctx.Err()
		}

		logger.LogDebug(consts.DebugMessage, fmt.Sprintf("blockID for confirmation: %d", blockID))

		block := model.Block{}
		err := block.GetBlock(blockID)
		if err != nil {
			logger.LogError(consts.DBError, err)
			return err
		}
		hashStr := string(converter.BinToHex(block.Hash))
		logger.LogDebug(consts.DebugMessage, fmt.Sprintf("hash for check: %x", hashStr))
		if len(hashStr) == 0 {
			logger.LogDebug(consts.DebugMessage, "len(hash) == 0")
			continue
		}

		var hosts []string
		if config.ConfigIni["test_mode"] == "1" {
			hosts = []string{"localhost"}
		} else {
			hosts, err = model.GetFullNodesHosts()
			if err != nil {
				logger.LogError(consts.DBError, err)
				return err
			}
		}

		ch := make(chan string)
		for i := 0; i < len(hosts); i++ {
			// TODO: ports should be in the table hosts
			host := hosts[i] + ":" + utils.GetTcpPort(config.ConfigIni)
			logger.LogDebug(consts.DebugMessage, fmt.Sprintf("host %v", host))
			go func() {
				IsReachable(host, blockID, ch)
			}()
		}
		var answer string
		var st0, st1 int64
		for i := 0; i < len(hosts); i++ {
			answer = <-ch
			logger.LogDebug(consts.DebugMessage, fmt.Sprintf("answer == hash, (%s = %s)", answer, hashStr))
			if answer == hashStr {
				st1++
			} else {
				st0++
			}
			logger.LogDebug(consts.DebugMessage, fmt.Sprintf("st0 %v  st1 %v", st0, st1))
		}
		confirmation := &model.Confirmation{}
		err = confirmation.GetConfirmation(blockID)
		if err == nil {
			confirmation.BlockID = blockID
			confirmation.Good = int32(st1)
			confirmation.Bad = int32(st0)
			confirmation.Time = int32(time.Now().Unix())
			err = confirmation.Save()
			if err != nil {
				logger.LogError(consts.DBError, err)
				return err
			}
		} else {
			confirmation.BlockID = blockID
			confirmation.Good = int32(st1)
			confirmation.Bad = int32(st0)
			confirmation.Time = int32(time.Now().Unix())
			err = confirmation.Save()
			if err != nil {
				logger.LogError(consts.DBError, err)
				return err
			}
		}
		logger.LogDebug(consts.DebugMessage, fmt.Sprintf("blockID > startBlockID && st1 >= consts.MIN_CONFIRMED_NODES %d>%d && %d>=%d\n", blockID, startBlockID, st1, consts.MIN_CONFIRMED_NODES))
		if blockID > startBlockID && st1 >= consts.MIN_CONFIRMED_NODES {
			break
		}
	}

	return nil

}

func checkConf(host string, blockID int64) string {
	logger.LogDebug(consts.FuncStarted, fmt.Sprintf("confirmation: check conf for host: %s", host))
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		logger.LogDebug(consts.ConnectionError, err)
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
		logger.LogError(consts.ConnectionError, fmt.Sprintf("send request error: %v", err))
		return "0"
	}

	resp := &tcpserver.ConfirmResponse{}
	err = tcpserver.ReadRequest(resp, conn)
	if err != nil {
		logger.LogError(consts.ConnectionError, fmt.Sprintf("read request error: %v", err))
		return "0"
	}
	return string(converter.BinToHex(resp.Hash))
}

// IsReachable checks if there is blockID on the host
func IsReachable(host string, blockID int64, ch0 chan string) {
	logger.LogDebug(consts.FuncStarted, fmt.Sprintf("IsReachable %v", host))
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
