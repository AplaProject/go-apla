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

	log "github.com/sirupsen/logrus"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
)

const BlocksPerRequest int32 = 1000

// Type7 writes the body of the specified block
// blocksCollection and queue_parser_blocks daemons send the request through p.GetBlocks()
func Type7(request *GetBodiesRequest, w net.Conn) error {
	block := &model.Block{}

	var blocks []model.Block
	var err error
	if request.ReverseOrder {
		blocks, err = block.GetReverseBlockchain(int64(request.BlockID), BlocksPerRequest)
	} else {
		blocks, err = block.GetBlocksFrom(int64(request.BlockID-1), "ASC", BlocksPerRequest)
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_id": request.BlockID}).Error("Error getting 1000 blocks from block_id")
		return err
	}

	if len(blocks) == 0 {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_id": request.BlockID}).Warn("Requesting nonexistent blocks from block_id")
		return err
	}

	for _, b := range blocks {
		if err := SendRequest(&GetBodyResponse{Data: b.Data}, w); err != nil {
			return err
		}
	}

	return nil
}
