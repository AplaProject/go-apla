// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package tcpserver

import (
	"net"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/network"
	"github.com/AplaProject/go-apla/packages/service"

	log "github.com/sirupsen/logrus"
)

// HandleTCPRequest proceed TCP requests
func HandleTCPRequest(rw net.Conn) {
	dType := &network.RequestType{}
	err := dType.Read(rw)
	if err != nil {
		log.Errorf("read request type failed: %s", err)
		return
	}

	log.WithFields(log.Fields{"request_type": dType.Type}).Debug("tcpserver got request type")
	var response interface{}

	switch dType.Type {
	case network.RequestTypeFullNode:
		if service.IsNodePaused() {
			return
		}
		err = Type1(rw)

	case network.RequestTypeNotFullNode:
		if service.IsNodePaused() {
			return
		}
		response, err = Type2(rw)

	case network.RequestTypeStopNetwork:
		req := &network.StopNetworkRequest{}
		if err = req.Read(rw); err == nil {
			err = Type3(req, rw)
		}

	case network.RequestTypeConfirmation:
		if service.IsNodePaused() {
			return
		}

		req := &network.ConfirmRequest{}
		if err = req.Read(rw); err == nil {
			response, err = Type4(req)
		}

	case network.RequestTypeBlockCollection:
		req := &network.GetBodiesRequest{}
		if err = req.Read(rw); err == nil {
			err = Type7(req, rw)
		}

	case network.RequestTypeMaxBlock:
		response, err = Type10()
	}

	if err != nil || response == nil {
		return
	}

	log.WithFields(log.Fields{"response": response, "request_type": dType.Type}).Debug("tcpserver responded")
	if err = response.(network.SelfReaderWriter).Write(rw); err != nil {
		// err = SendRequest(response, rw)
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
