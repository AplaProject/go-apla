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
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/network"

	log "github.com/sirupsen/logrus"
)

// Type4 writes the hash of the specified block
// The request is sent by 'confirmations' daemon
func Type4(r *network.ConfirmRequest) (*network.ConfirmResponse, error) {
	resp := &network.ConfirmResponse{}
	block := &model.Block{}
	found, err := block.Get(int64(r.BlockID))
	if err != nil || !found {
		hash := [32]byte{}
		resp.Hash = hash[:]
	} else {
		resp.Hash = block.Hash // can we send binary data ?
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_id": r.BlockID}).Error("Getting block")
	} else if len(block.Hash) == 0 {
		log.WithFields(log.Fields{"type": consts.DBError, "block_id": r.BlockID}).Warning("Block not found")
	}
	return resp, nil
}
