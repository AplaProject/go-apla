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
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// BlocksPerRequest contains count of blocks per request
const BlocksPerRequest int32 = 1000

// Type7 writes the body of the specified block
// blocksCollection and queue_parser_blocks daemons send the request through p.GetBlocks()
func Type7(request *GetBodiesRequest, w net.Conn) error {
	var blocks []*blockchain.Block
	var err error
	order := 1
	if request.ReverseOrder {
		order = -1
	}
	blocks, err = blockchain.GetNBlocksFrom(request.BlockHash, int(BlocksPerRequest), order)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_hash": request.BlockHash}).Error("Error getting 1000 blocks from block_hash")
		return err
	}

	if len(blocks) == 0 {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_hash": request.BlockHash}).Warn("Requesting nonexistent blocks from block_hash")
		return err
	}

	nodePrivateKey, _, err := utils.GetNodeKeys()
	if err != nil {
		return err
	}
	for _, b := range blocks {
		data, err := b.Marshal(nodePrivateKey)
		if err != nil {
			return err
		}
		if err := SendRequest(&GetBodyResponse{Data: data}, w); err != nil {
			return err
		}
	}

	return nil
}
