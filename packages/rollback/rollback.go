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

package rollback

import (
	"bytes"
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// ToBlockID rollbacks blocks till blockID
func ToBlockID(blockID int64, dbTransaction *model.DbTransaction, logger *log.Entry) error {
	_, err := model.MarkVerifiedAndNotUsedTransactionsUnverified()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("marking verified and not used transactions unverified")
		return err
	}

	limit := 1000
	// roll back our blocks
	for {
		block := &model.Block{}
		blocks, err := block.GetBlocks(blockID, int32(limit))
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting blocks")
			return err
		}
		if len(blocks) == 0 {
			break
		}
		for _, block := range blocks {
			// roll back our blocks to the block blockID
			err = RollbackBlock(block.Data, true)
			if err != nil {
				return err
			}
		}
		blocks = blocks[:0]
	}
	block := &model.Block{}
	_, err = block.Get(blockID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block")
		return err
	}

	isFirstBlock := blockID == 1
	header, err := utils.ParseBlockHeader(bytes.NewBuffer(block.Data), !isFirstBlock)
	if err != nil {
		return err
	}

	ib := &model.InfoBlock{
		Hash:           block.Hash,
		BlockID:        header.BlockID,
		Time:           header.Time,
		EcosystemID:    header.EcosystemID,
		KeyID:          header.KeyID,
		NodePosition:   converter.Int64ToStr(header.NodePosition),
		CurrentVersion: strconv.Itoa(header.Version),
	}

	err = ib.Update(dbTransaction)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating info block")
		return err
	}

	return nil
}
