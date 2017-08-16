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
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"

	"context"
)

/* Берем блок. Если блок имеет лучший хэш, то ищем, в каком блоке у нас пошла вилка // Take the block. If the block has the best hash, then look for the block where the fork started
 * Если вилка пошла менее чем variables->rollback_blocks блоков назад, то // If the fork begins less then variables->rollback_blocks blocks ago, than
 *  - получаем всю цепочку блоков, // get the whole chain of blocks
 *  - откатываем фронтальные данные от наших блоков, // roll back the frontal data from our blocks
 *  - заносим фронт. данные из новой цепочки // insert the frontal data from a new chain
 *  - если нет ошибок, то откатываем наши данные из блоков // if there is no error, then roll back our data from the blocks
 *  - и заносим новые данные // and insert new data
 *  - если где-то есть ошибки, то откатываемся к нашим прежним данным // if there are errors, then roll back to the former data
 * Если вилка была давно, то ничего не трогаем, и оставлеяем скрипту blocks_collection.php // if the fork was long ago then do not touch anything and leave the script blocks_collection.php
 * Ограничение variables->rollback_blocks нужно для защиты от подставных блоков // the limitation variables->rollback_blocks is needed for the protection against the false blocks
 *
 * */

// QueueParserBlocks parses blocks from the queue
func QueueParserBlocks(d *daemon, ctx context.Context) error {

	locked, err := DbLock(ctx, d.goRoutineName)
	if !locked || err != nil {
		return err
	}
	defer DbUnlock(d.goRoutineName)

	infoBlock := &model.InfoBlock{}
	err = infoBlock.GetInfoBlock()
	if err != nil {
		return err
	}
	queueBlock := &model.QueueBlock{}
	err = queueBlock.GetQueueBlock()
	if err != nil {
		return err
	}
	if len(queueBlock.Hash) == 0 {
		return err
	}
	queueBlock.Hash = converter.BinToHex(queueBlock.Hash)
	infoBlock.Hash = converter.BinToHex(infoBlock.Hash)

	// check if the block gets in the rollback_blocks_1 limit
	if queueBlock.BlockID > infoBlock.BlockID+consts.RB_BLOCKS_1 {
		queueBlock.Delete()
		return utils.ErrInfo("rollback_blocks_1")
	}

	// is it old block in queue ?
	if queueBlock.BlockID <= infoBlock.BlockID {
		queueBlock.Delete()
		return utils.ErrInfo("old block")
	}

	// download blocks for check
	fullNode := &model.FullNode{}

	err = fullNode.FindNodeByID(queueBlock.FullNodeID)
	if err != nil {
		queueBlock.Delete()
		return utils.ErrInfo(err)
	}

	blockID := queueBlock.BlockID

	p := new(parser.Parser)
	p.GoroutineName = d.goRoutineName

	host := GetHostPort(fullNode.Host)
	err = p.GetBlocks(blockID, host, "rollback_blocks_1", d.goRoutineName, 7)
	if err != nil {
		log.Error("v", err)
		queueBlock.Delete()
		return utils.ErrInfo(err)
	}
	return nil
}
