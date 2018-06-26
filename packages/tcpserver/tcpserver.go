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

package tcpserver

import (
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/service"

	log "github.com/sirupsen/logrus"
)

var (
	counter int64
)

// HandleTCPRequest proceed TCP requests
func HandleTCPRequest(rw net.Conn) {
	defer func() {
		atomic.AddInt64(&counter, -1)
	}()

	count := atomic.AddInt64(&counter, +1)
	if count > 20 {
		return
	}

	dType := &RequestType{}
	err := dType.Read(rw)
	if err != nil {
		log.Errorf("read request type failed: %s", err)
		return
	}

	log.WithFields(log.Fields{"request_type": dType.Type}).Debug("tcpserver got request type")
	var response interface{}

	switch dType.Type {
	case RequestTypeFullNode:
		if service.IsNodePaused() {
			return
		}
		err = Type1(rw)

	case RequestTypeNotFullNode:
		if service.IsNodePaused() {
			return
		}
		response, err = Type2(rw)

	case RequestTypeStopNetwork:
		req := &StopNetworkRequest{}
		if err = ReadRequest(req, rw); err == nil {
			err = Type3(req, rw)
		}

	case RequestTypeConfirmation:
		if service.IsNodePaused() {
			return
		}
		req := &ConfirmRequest{}
		err = ReadRequest(req, rw)
		if err == nil {
			response, err = Type4(req)
		}

	case RequestTypeBlockCollection:
		req := &GetBodiesRequest{}
		err = ReadRequest(req, rw)
		if err == nil {
			err = Type7(req, rw)
		}

	case RequestTypeMaxBlock:
		response, err = Type10()
	}

	if err != nil || response == nil {
		return
	}

	log.WithFields(log.Fields{"response": response, "request_type": dType.Type}).Debug("tcpserver responded")
	err = SendRequest(response, rw)
	if err != nil {
		log.Errorf("tcpserver handle error: %s", err)
	}
}

// TcpListener is listening tcp address
func TcpListener(laddr string) error {

	if strings.HasPrefix(laddr, "127.") {
		log.Warn("Listening at local address: ", laddr)
	}

	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": laddr}).Error("Error listening")
		return err
	}

	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": laddr}).Error("Error accepting")
				time.Sleep(time.Second)
			} else {
				go func(conn net.Conn) {
					HandleTCPRequest(conn)
					conn.Close()
				}(conn)
			}
		}
	}()

	return nil
}
