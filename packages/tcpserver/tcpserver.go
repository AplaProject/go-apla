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
	"flag"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"

	log "github.com/sirupsen/logrus"
)

var (
	counter int64
)

func init() {
	flag.Parse()
}

// HandleTCPRequest proceed TCP requests
func HandleTCPRequest(rw io.ReadWriter) {
	defer func() {
		atomic.AddInt64(&counter, -1)
	}()

	count := atomic.AddInt64(&counter, +1)
	if count > 20 {
		return
	}

	dType := &TransactionType{}
	err := ReadRequest(dType, rw)
	if err != nil {
		log.Errorf("read request type failed: %s", err)
		return
	}

	log.WithFields(log.Fields{"request_type": dType.Type}).Debug("tcpserver got request type")
	var response interface{}

	switch dType.Type {
	case 1:
		req := &DisRequest{}
		err = ReadRequest(req, rw)
		if err == nil {
			err = Type1(req, rw)
		}

	case 2:
		req := &DisRequest{}
		err = ReadRequest(req, rw)
		if err == nil {
			response, err = Type2(req)
		}

	case 4:
		req := &ConfirmRequest{}
		err = ReadRequest(req, rw)
		if err == nil {
			response, err = Type4(req)
		}

	case 7:
		req := &GetBodyRequest{}
		err = ReadRequest(req, rw)
		if err == nil {
			response, err = Type7(req)
		}

	case 10:
		response, err = Type10()
	}

	if err != nil {
		return
	}
	if response == nil {
		return
	}

	log.WithFields(log.Fields{"response": response}).Debug("tcpserver responded")
	err = SendRequest(response, rw)
	if err != nil {
		log.Errorf("tcpserver handle error: %s", err)
	}
}

func TcpListener(laddr string) error {
	l, err := net.Listen("tcp4", laddr)
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
