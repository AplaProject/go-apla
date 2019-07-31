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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/network"
	log "github.com/sirupsen/logrus"
)

// Type7 writes the body of the specified block
// blocksCollection and queue_parser_blocks daemons send the request through p.GetBlocks()
func Type7(request *network.GetBodiesRequest, w net.Conn) error {
	block := &model.Block{}

	var blocks []model.Block
	var err error
	if request.ReverseOrder {
		blocks, err = block.GetReverseBlockchain(int64(request.BlockID), network.BlocksPerRequest)
	} else {
		blocks, err = block.GetBlocksFrom(int64(request.BlockID-1), "ASC", network.BlocksPerRequest)
	}

	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_id": request.BlockID}).Error("Error getting 1000 blocks from block_id")
		if err := network.WriteInt(0, w); err != nil {
			log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending 0 requested blocks")
		}
		return err
	}

	if err := network.WriteInt(int64(len(blocks)), w); err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending requested blocks count")
		return err
	}

	if err := network.WriteInt(lenOfBlockData(blocks), w); err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending requested blocks data length")
		return err
	}

	for _, b := range blocks {
		br := &network.GetBodyResponse{Data: b.Data}
		if err := br.Write(w); err != nil {
			return err
		}
	}

	return nil
}

func lenOfBlockData(blocks []model.Block) int64 {
	var length int64
	for i := 0; i < len(blocks); i++ {
		length += int64(len(blocks[i].Data))
	}

	return length
}
