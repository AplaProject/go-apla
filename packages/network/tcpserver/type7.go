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

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/network"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

const BlocksPerRequest = 5

// Type7 writes the body of the specified block
// blocksCollection and queue_parser_blocks daemons send the request through p.GetBlocks()
func Type7(request *network.GetBodiesRequest, w net.Conn) error {
	var blocks []*blockchain.BlockWithHash
	var err error
	order := 1
	if request.ReverseOrder {
		order = -1
	}
	blocks, err = blockchain.GetNBlocksFrom(request.BlockHash, int(BlocksPerRequest), order)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_hash": request.BlockHash}).Error("Error getting 1000 blocks from block_hash")
		if err := network.WriteInt(0, w); err != nil {
			log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending 0 requested blocks")
		}
		return err
	}

	if len(blocks) == 0 {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_hash": request.BlockHash}).Warn("Requesting nonexistent blocks from block_hash")
		return nil
	}
	if err := network.WriteInt(int64(len(blocks)), w); err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending requested blocks count")
		return err
	}

	nodePrivateKey, _, err := utils.GetNodeKeys()
	if err != nil {
		return err
	}
	for _, b := range blocks {
		data, err := b.Block.Marshal(nodePrivateKey)
		if err != nil {
			return err
		}
		br := &network.GetBodyResponse{Data: data}
		if err := br.Write(w); err != nil {
			return err
		}
	}

	return nil
}
