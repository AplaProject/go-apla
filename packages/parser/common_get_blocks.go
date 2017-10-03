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

package parser

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func GetBlocks(blockID int64, host string, rollbackBlocks string, dataTypeBlockBody int64) error {
	rollback := consts.RB_BLOCKS_1
	if rollbackBlocks == "rollback_blocks_2" {
		rollback = consts.RB_BLOCKS_2
	}

	config := &model.Config{}
	err := config.GetConfig()
	if err != nil {
		return utils.ErrInfo(err)
	}

	badBlocks := make(map[int64]string)
	if len(config.BadBlocks) > 0 {
		err = json.Unmarshal([]byte(config.BadBlocks), &badBlocks)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	blocks := make([]*Block, 0)
	var count int64

	for {
		if blockID < 2 {
			return utils.ErrInfo(errors.New("block_id < 2"))
		}
		// if the limit of blocks received from the node was exaggerated
		if count > int64(rollback) {
			return utils.ErrInfo(errors.New("count > variables[rollback_blocks]"))
		}

		// load the block body from the host
		binaryBlock, err := utils.GetBlockBody(host, blockID, dataTypeBlockBody)
		if err != nil {
			return utils.ErrInfo(err)
		}

		block, err := ProcessBlock(binaryBlock)
		if err != nil {
			return utils.ErrInfo(err)
		}

		if badBlocks[block.Header.BlockID] == string(converter.BinToHex(block.Header.Sign)) {
			return utils.ErrInfo(errors.New("bad block"))
		}
		if block.Header.BlockID != blockID {
			return utils.ErrInfo(errors.New("bad block_data['block_id']"))
		}

		// save the block
		blocks = append(blocks, block)
		blockID--
		count++

		// the public key of the one who has generated this block
		nodePublicKey, err := GetNodePublicKeyWalletOrCB(block.Header.WalletID, block.Header.StateID)
		if err != nil {
			return utils.ErrInfo(err)
		}

		// SIGN from 128 bytes to 512 bytes. Signature of TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT
		forSign := fmt.Sprintf("0,%v,%x,%v,%v,%v,%s", block.Header.BlockID, block.PrevHeader.Hash, block.Header.Time,
			block.Header.WalletID, block.Header.StateID, block.MrklRoot)

		// check the signature
		_, okSignErr := utils.CheckSign([][]byte{nodePublicKey}, forSign, block.Header.Sign, true)
		if okSignErr == nil {
			// this block is matched with our blockchain
			break
		}
	}

	// mark all transaction as unverified
	_, err = model.MarkVerifiedAndNotUsedTransactionsUnverified()
	if err != nil {
		return utils.ErrInfo(err)
	}

	// we have the slice of blocks for applying
	// first of all we should rollback old blocks
	block := &model.Block{}
	myRollbackBlocks, err := block.GetBlocksFrom(blockID, "desc")
	if err != nil {
		return utils.ErrInfo(err)
	}
	for _, block := range myRollbackBlocks {
		log.Debug("We roll away blocks before plug", blockID)
		err := BlockRollback(block.Data)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	dbTransaction, err := model.StartTransaction()
	if err != nil {
		return utils.ErrInfo(err)
	}

	// go through new blocks in reverse order
	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]

		// our blockchain is changing, so we should read again previous block
		err := block.readPreviousBlock()
		if err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}

		if err := block.CheckBlock(); err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}

		if err := block.playBlock(dbTransaction); err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}

		// for last block we should update block info
		if i == 0 {
			err := UpdBlockInfo(dbTransaction, block)
			if err != nil {
				dbTransaction.Rollback()
				return utils.ErrInfo(err)
			}
		}
	}

	err = dbTransaction.Commit()
	return err
}
