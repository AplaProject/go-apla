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

package daemons

import (
	"context"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/queue"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

/* Take the block from the queue. If this block has the bigger block id than the last block from our chain, then find the fork
 * If fork begins less then variables->rollback_blocks blocks ago, than
 *  - get the whole chain of blocks
 *  - roll back data from our blocks
 *  - insert the frontal data from a new chain
 *  - if there is no error, then roll back our data from the blocks
 *  - and insert new data
 *  - if there are errors, then roll back to the former data
 * */

// QueueParserBlocks parses and applies blocks from the queue
func QueueParserBlocks(ctx context.Context, d *daemon) error {
	DBLock()
	defer DBUnlock()

	infoBlock, _, found, err := blockchain.GetLastBlock()
	if !found {
		return nil
	}
	queueBlock, err := queue.ValidateBlockQueue.Dequeue()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("getting block from validate queue")
		return err
	}
	// check if the block gets in the rollback_blocks_1 limit
	if queueBlock.BlockID > infoBlock.Header.BlockID+syspar.GetRbBlocks1() {
		return utils.ErrInfo("rollback_blocks_1")
	}

	// is it old block in queue ?
	if queueBlock.BlockID <= infoBlock.Header.BlockID {
		return utils.ErrInfo(fmt.Errorf("old block %d <= %d", queueBlock.BlockID, infoBlock.Header.BlockID))
	}
	if queueBlock.FullNodeID == conf.Config.KeyID {
		d.logger.WithFields(log.Fields{"type": consts.DuplicateObject}).Debug("queueBlock generated by myself", queueBlock.BlockID)
		return utils.ErrInfo(fmt.Errorf("queueBlock generated by myself: %d", queueBlock.BlockID))
	}

	nodeHost, err := syspar.GetNodeHostByPosition(queueBlock.FullNodeID)
	if err != nil {
		return utils.ErrInfo(err)
	}
	blockID := queueBlock.BlockID

	host := utils.GetHostPort(nodeHost)
	// update our chain till maxBlockID from the host
	return UpdateChain(ctx, d, host, blockID)
}
