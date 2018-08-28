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
	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/network"

	log "github.com/sirupsen/logrus"
)

// Type4 writes the hash of the specified block
// The request is sent by 'confirmations' daemon
func Type4(r *network.ConfirmRequest) (*network.ConfirmResponse, error) {
	resp := &network.ConfirmResponse{}
	block, found, err := blockchain.GetBlock(r.BlockHash)
	if err != nil || !found {
		hash := [32]byte{}
		resp.Hash = hash[:]
	} else {
		resp.Hash = block.Header.Hash // can we send binary data ?
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_hash": r.BlockHash}).Error("Getting block")
	} else if !found {
		log.WithFields(log.Fields{"type": consts.DBError, "block_hash": r.BlockHash}).Warning("Block not found")
	}
	return resp, nil
}
