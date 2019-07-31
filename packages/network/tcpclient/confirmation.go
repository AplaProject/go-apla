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

package tcpclient

import (
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/network"
	log "github.com/sirupsen/logrus"
)

func CheckConfirmation(host string, blockID int64, logger *log.Entry) (hash string) {
	conn, err := newConnection(host)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": host, "block_id": blockID}).Debug("dialing to host")
		return "0"
	}
	defer conn.Close()

	rt := &network.RequestType{Type: network.RequestTypeConfirmation}
	if err = rt.Write(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_id": blockID}).Error("sending request type")
		return "0"
	}

	req := &network.ConfirmRequest{
		BlockID: uint32(blockID),
	}
	if err = req.Write(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_id": blockID}).Error("sending confirmation request")
		return "0"
	}

	resp := &network.ConfirmResponse{}

	if err := resp.Read(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_id": blockID}).Error("receiving confirmation response")
		return "0"
	}
	return string(converter.BinToHex(resp.Hash))
}
