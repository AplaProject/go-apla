// MIT License
//
// Copyright (c) 2016 GenesisCommunity
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package tcpserver

import (
	"net"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

// BlocksPerRequest contains count of blocks per request
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
