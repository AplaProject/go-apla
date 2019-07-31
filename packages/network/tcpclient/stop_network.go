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
	"github.com/AplaProject/go-apla/packages/network"
)

func SendStopNetwork(addr string, req *network.StopNetworkRequest) error {
	conn, err := newConnection(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	rt := &network.RequestType{
		Type: network.RequestTypeStopNetwork,
	}

	if err = rt.Write(conn); err != nil {
		return err
	}

	if err = req.Write(conn); err != nil {
		return err
	}

	res := &network.StopNetworkResponse{}
	if err = res.Read(conn); err != nil {
		return err
	}

	if len(res.Hash) != consts.HashSize {
		return network.ErrNotAccepted
	}

	return nil
}
