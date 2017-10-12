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
	//	"fmt"

	//	"runtime"

	"sync/atomic"

	"io"

	"net"
	"time"

	"github.com/op/go-logging"
)

var (
	log     = logging.MustGetLogger("tcpserver")
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

	log.Debugf("tcpservers: got request type: %d", dType.Type)
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
		log.Errorf("tcpserver: parse request error: %s", err)
		return
	}
	if response == nil {
		return
	}

	log.Debugf("tcpserver response: %+v", response)
	err = SendRequest(response, rw)
	if err != nil {
		log.Errorf("tcpserver handle error: %s", err)
	}
}

func TcpListener(laddr string) error {
	log.Debugf("listen addres: %s", laddr)

	l, err := net.Listen("tcp4", laddr)
	if err != nil {
		log.Error("Error listening:", err)
		return err
	}

	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Error("Error accepting:", err)
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
