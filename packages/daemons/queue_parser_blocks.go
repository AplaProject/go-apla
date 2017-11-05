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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/utils"

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
func QueueParserBlocks(d *daemon, ctx context.Context) error {
	DBLock()
	defer DBUnlock()

	infoBlock := &model.InfoBlock{}
	_, err := infoBlock.Get()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting info block")
		return err
	}
	queueBlock := &model.QueueBlock{}
	_, err = queueBlock.Get()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting queue block")
		return err
	}
	if len(queueBlock.Hash) == 0 {
		d.logger.WithFields(log.Fields{"type": consts.NotFound}).Debug("queue block not found")
		return err
	}

	// check if the block gets in the rollback_blocks_1 limit
	if queueBlock.BlockID > infoBlock.BlockID+consts.RB_BLOCKS_1 {
		err = queueBlock.Delete()
		if err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting queue block")
		}
		return utils.ErrInfo("rollback_blocks_1")
	}

	// is it old block in queue ?
	if queueBlock.BlockID <= infoBlock.BlockID {
		err = queueBlock.Delete()
		if err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting queue block")
		}
		return utils.ErrInfo("old block")
	}

	// download blocks for check
	fullNode := &model.FullNode{}

	_, err = fullNode.FindNodeByID(queueBlock.FullNodeID)
	if err != nil {
		err = queueBlock.Delete()
		if err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting queue block")
		}
		return utils.ErrInfo(err)
	}

	blockID := queueBlock.BlockID

	host := getHostPort(fullNode.Host)
	err = parser.GetBlocks(blockID, host, "rollback_blocks_1", 7)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback_blocks")
		err = queueBlock.Delete()
		if err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting queue block")
		}
		return utils.ErrInfo(err)
	}
	return nil
}
