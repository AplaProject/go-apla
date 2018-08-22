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
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/tcpserver"

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

	hashes, err := blockchain.GetUnconfirmedBlocks(consts.MIN_CONFIRMED_NODES)
	if err != nil {
		return err
	}
	if len(hashes) == 0 {
		return nil
	}

	return confirmationsBlocks(ctx, d, hashes)
}

func confirmationsBlocks(ctx context.Context, d *daemon, blocks []*blockchain.Block) error {
	for _, block := range blocks {
		if err := ctx.Err(); err != nil {
			d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": err}).Error("error in context")
			return err
		}

		hashStr := string(converter.BinToHex(block.Header.Hash))
		d.logger.WithFields(log.Fields{"hash": hashStr}).Debug("checking hash")

		hosts, err := filterBannedHosts(syspar.GetRemoteHosts())
		if err != nil {
			return err
		}

		ch := make(chan string)
		for i := 0; i < len(hosts); i++ {
			host, err := NormalizeHostAddress(hosts[i], consts.DEFAULT_TCP_PORT)
			if err != nil {
				d.logger.WithFields(log.Fields{"host": host[i], "type": consts.ParseError, "error": err}).Error("wrong host address")
				continue
			}

			d.logger.WithFields(log.Fields{"host": host, "block_hash": block.Header.Hash}).Debug("checking block id confirmed at node")
			go func() {
				IsReachable(host, block.Header.Hash, ch, d.logger)
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
		confirmation.BlockID = block.Header.BlockID
		confirmation.Good = st1
		confirmation.Bad = st0
		confirmation.Time = time.Now().Unix()
		if err = blockchain.InsertConfirmation(block.Header.Hash, confirmation); err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("saving confirmation")
			return err
		}
	}

	return nil
}

func checkConf(host string, blockHash []byte, logger *log.Entry) string {
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": host, "block_hash": blockHash}).Debug("dialing to host")
		return "0"
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(consts.READ_TIMEOUT * time.Second))
	conn.SetWriteDeadline(time.Now().Add(consts.WRITE_TIMEOUT * time.Second))

	if err = tcpserver.SendRequestType(tcpserver.RequestTypeConfirmation, conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_hash": blockHash}).Error("sending request type")
		return "0"
	}

	req := &tcpserver.ConfirmRequest{
		BlockHash: blockHash,
	}
	if err = tcpserver.SendRequest(req, conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_hash": blockHash}).Error("sending confirmation request")
		return "0"
	}

	resp := &tcpserver.ConfirmResponse{}
	err = tcpserver.ReadRequest(resp, conn)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_hash": blockHash}).Error("receiving confirmation response")
		return "0"
	}
	return string(converter.BinToHex(resp.Hash))
}

// IsReachable checks if there is blockID on the host
func IsReachable(host string, blockHash []byte, ch0 chan string, logger *log.Entry) {
	ch := make(chan string, 1)
	go func() {
		ch <- checkConf(host, blockHash, logger)
	}()
	select {
	case reachable := <-ch:
		ch0 <- reachable
	case <-time.After(consts.WAIT_CONFIRMED_NODES * time.Second):
		ch0 <- "0"
	}
}

// NormalizeHostAddress get address. if port not defined returns combined string with ip and defaultPort
func NormalizeHostAddress(address string, defaultPort int) (string, error) {

	_, _, err := net.SplitHostPort(address)
	if err != nil {
		if strings.HasSuffix(err.Error(), "missing port in address") {
			return fmt.Sprintf("%s:%d", address, defaultPort), nil
		}

		return "", err
	}

	return address, nil
}
