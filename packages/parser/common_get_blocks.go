// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
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

package parser

import (
	"errors"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// GetBlocks is returning blocks
func GetBlocks(blockID int64, host string) error {
	rollback := syspar.GetRbBlocks1()

	badBlocks := make(map[int64]string)

	blocks := make([]*Block, 0)
	var count int64

	for {
		if blockID < 2 {
			log.WithFields(log.Fields{"type": consts.BlockIsFirst}).Error("block id is smaller than 2")
			return utils.ErrInfo(errors.New("block_id < 2"))
		}
		// if the limit of blocks received from the node was exaggerated
		if count > int64(rollback) {
			log.WithFields(log.Fields{"count": count, "max_count": int64(rollback)}).Error("limit of received from the node was exaggerated")
			return utils.ErrInfo(errors.New("count > variables[rollback_blocks]"))
		}

		// load the block body from the host
		binaryBlock, err := utils.GetBlockBody(host, blockID, consts.DATA_TYPE_BLOCK_BODY)
		if err != nil {
			return utils.ErrInfo(err)
		}

		block, err := ProcessBlockWherePrevFromBlockchainTable(binaryBlock)
		if err != nil {
			return utils.ErrInfo(err)
		}

		if badBlocks[block.Header.BlockID] == string(converter.BinToHex(block.Header.Sign)) {
			log.WithFields(log.Fields{"block_id": block.Header.BlockID, "type": consts.InvalidObject}).Error("block is bad")
			return utils.ErrInfo(errors.New("bad block"))
		}
		if block.Header.BlockID != blockID {
			log.WithFields(log.Fields{"header_block_id": block.Header.BlockID, "block_id": blockID, "type": consts.InvalidObject}).Error("block ids does not match")
			return utils.ErrInfo(errors.New("bad block_data['block_id']"))
		}

		// TODO: add checking for MAX_BLOCK_SIZE

		// the public key of the one who has generated this block
		nodePublicKey, err := syspar.GetNodePublicKeyByPosition(block.Header.NodePosition)
		if err != nil {
			log.WithFields(log.Fields{"header_block_id": block.Header.BlockID, "block_id": blockID, "type": consts.InvalidObject}).Error("block ids does not match")
			return utils.ErrInfo(err)
		}

		// SIGN from 128 bytes to 512 bytes. Signature of TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT
		forSign := fmt.Sprintf("0,%v,%x,%v,%v,%v,%v,%s",
			block.Header.BlockID, block.PrevHeader.Hash, block.Header.Time,
			block.Header.EcosystemID, block.Header.KeyID, block.Header.NodePosition,
			block.MrklRoot,
		)

		// save the block
		blocks = append(blocks, block)
		blockID--
		count++

		// check the signature
		_, okSignErr := utils.CheckSign([][]byte{nodePublicKey}, forSign, block.Header.Sign, true)
		if okSignErr == nil {
			break
		}
	}

	// mark all transaction as unverified
	_, err := model.MarkVerifiedAndNotUsedTransactionsUnverified()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"type":  consts.DBError,
		}).Error("marking verified and not used transactions unverified")
		return utils.ErrInfo(err)
	}

	// we have the slice of blocks for applying
	// first of all we should rollback old blocks
	block := &model.Block{}
	myRollbackBlocks, err := block.GetBlocksFrom(blockID, "desc")
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting rollback blocks from blockID")
		return utils.ErrInfo(err)
	}
	for _, block := range myRollbackBlocks {
		err := RollbackTxFromBlock(block.Data)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("starting transaction")
		return utils.ErrInfo(err)
	}

	// go through new blocks from the smallest block_id to the largest block_id
	prevBlocks := make(map[int64]*Block, 0)

	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]

		if prevBlocks[block.Header.BlockID-1] != nil {
			block.PrevHeader.Hash = prevBlocks[block.Header.BlockID-1].Header.Hash
			block.PrevHeader.Time = prevBlocks[block.Header.BlockID-1].Header.Time
			block.PrevHeader.BlockID = prevBlocks[block.Header.BlockID-1].Header.BlockID
			block.PrevHeader.EcosystemID = prevBlocks[block.Header.BlockID-1].Header.EcosystemID
			block.PrevHeader.KeyID = prevBlocks[block.Header.BlockID-1].Header.KeyID
			block.PrevHeader.NodePosition = prevBlocks[block.Header.BlockID-1].Header.NodePosition
		}

		forSha := fmt.Sprintf("%d,%x,%s,%d,%d,%d,%d", block.Header.BlockID, block.PrevHeader.Hash, block.MrklRoot, block.Header.Time, block.Header.EcosystemID, block.Header.KeyID, block.Header.NodePosition)
		hash, err := crypto.DoubleHash([]byte(forSha))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("double hashing block")
		}
		block.Header.Hash = hash

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
		if block.SysUpdate {
			if err := syspar.SysUpdate(dbTransaction); err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
				return utils.ErrInfo(err)
			}
		}
	}

	// If all right we can delete old blockchain and write new
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
