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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/crypto"
)

func GetBlocks(blockID int64, host string, rollbackBlocks string, dataTypeBlockBody int64) error {
	rollback := consts.RB_BLOCKS_1
	if rollbackBlocks == "rollback_blocks_2" {
		rollback = consts.RB_BLOCKS_2
	}

	config := &model.Config{}
	_, err := config.Get()
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

		block, err := ProcessBlockWherePrevFromBlockchainTable(binaryBlock)
		if err != nil {
			return utils.ErrInfo(err)
		}

		if badBlocks[block.Header.BlockID] == string(converter.BinToHex(block.Header.Sign)) {
			return utils.ErrInfo(errors.New("bad block"))
		}
		if block.Header.BlockID != blockID {
			return utils.ErrInfo(errors.New("bad block_data['block_id']"))
		}

		// TODO: add checking for MAX_BLOCK_SIZE



		// the public key of the one who has generated this block
		nodePublicKey, err := GetNodePublicKeyWalletOrCB(block.Header.WalletID, block.Header.StateID)
		if err != nil {
			return utils.ErrInfo(err)
		}

		// SIGN from 128 bytes to 512 bytes. Signature of TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT
		forSign := fmt.Sprintf("0,%v,%x,%v,%v,%v,%s", block.Header.BlockID, block.PrevHeader.Hash, block.Header.Time, block.Header.WalletID, block.Header.StateID, block.MrklRoot)


		// save the block
		blocks = append(blocks, block)
		blockID--
		count++

		// check the signature
		_, okSignErr := utils.CheckSign([][]byte{nodePublicKey}, forSign, block.Header.Sign, true)
		if okSignErr == nil {
			// this block is matched with our blockchain
			log.Debug("this block is matched with our blockchain")
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
		err := RollbackTxFromBlock(block.Data)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	dbTransaction, err := model.StartTransaction()
	if err != nil {
		return utils.ErrInfo(err)
	}

	// go through new blocks from the smallest block_id to the largest block_id
	prevBlocks := make(map[int64]*Block, 0)

	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]

		log.Debug("i: %d / block: %v", i, block)

		if prevBlocks[block.Header.BlockID-1] != nil {
			log.Debug("prevBlock[intBlockId-1] != nil : %v", prevBlocks[block.Header.BlockID-1])
			log.Debug("prevBlock[intBlockId-1].Header.Hash : %x", prevBlocks[block.Header.BlockID-1].Header.Hash)
			block.PrevHeader.Hash = prevBlocks[block.Header.BlockID-1].Header.Hash
			block.PrevHeader.Time = prevBlocks[block.Header.BlockID-1].Header.Time
			block.PrevHeader.BlockID = prevBlocks[block.Header.BlockID-1].Header.BlockID
			block.PrevHeader.WalletID = prevBlocks[block.Header.BlockID-1].Header.WalletID
		}

		forSha := fmt.Sprintf("%d,%x,%s,%d,%d,%d", block.Header.BlockID, block.PrevHeader.Hash, block.MrklRoot, block.Header.Time, block.Header.WalletID, block.Header.StateID)
		log.Debug("block.Header.Time %v", block.Header.Time)
		log.Debug("block.PrevHeader.Time %v", block.PrevHeader.Time)

		hash, err := crypto.DoubleHash([]byte(forSha))
		if err != nil {
			log.Fatal(err)
		}
		block.Header.Hash = hash
		log.Debug("hash %x", hash)
		log.Debug("block.Header.Hash : %x", block.Header.Hash)

		if err := block.CheckBlock(); err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}

		if err := block.playBlock(dbTransaction); err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}
		prevBlocks[block.Header.BlockID] = block

		// for last block we should update block info
		if i == 0 {
			err := UpdBlockInfo(dbTransaction, block)
			if err != nil {
				dbTransaction.Rollback()
				return utils.ErrInfo(err)
			}
		}
	}

	// If all right we can delete old blockchain and write new
	log.Debug("If all right we can delete old blockchain and write new")
	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]
		// Delete old blocks from blockchain
		b := &model.Block{}
		err = b.DeleteById(dbTransaction, block.Header.BlockID)
		if err != nil {
			dbTransaction.Rollback()
			return err
		}
		// insert new blocks into blockchain
		if err := InsertIntoBlockchain(dbTransaction, block); err != nil {
			dbTransaction.Rollback()
			return err
		}
	}

	err = dbTransaction.Commit()
	return err
}
