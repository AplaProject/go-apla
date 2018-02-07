//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package tcpserver

import (
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GenesisCommunity/go-genesis/packages/consts"

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
	err := ReadRequest(dType, rw)
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

	if conf.Config.IsSupportingVDE() {
		return nil
	}

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
