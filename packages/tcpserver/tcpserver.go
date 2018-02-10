// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package tcpserver

import (
	"io"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"

	log "github.com/sirupsen/logrus"
)

var (
	counter int64
)

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

// TCPListener is listening tcp address
func TCPListener(addr string) error {

	if strings.HasPrefix(addr, "127.") {
		log.Warn("Listening at local address: ", addr)
	}

	l, err := net.Listen("tcp4", addr)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": addr}).Error("Error listening")
		return err
	}

	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": addr}).Error("Error accepting")
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
