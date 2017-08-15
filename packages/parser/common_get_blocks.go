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
	"io/ioutil"
	"os"

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/logging"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// GetOldBlocks gets previous blocks
// $get_block_script_name, $add_node_host is used only when working in protected mode and only from blocks_collection.php
func (p *Parser) GetOldBlocks(walletID, StateID, blockID int64, host string, goroutineName string, dataTypeBlockBody int64) error {
	log.Debug("walletId", walletID, "StateID", StateID, "blockID", blockID)
	err := p.GetBlocks(blockID, host, "rollback_blocks_2", goroutineName, dataTypeBlockBody)
	if err != nil {
		log.Error("v", err)
		return err
	}
	return nil
}

// GetBlocks gets blocks
func (p *Parser) GetBlocks(blockID int64, host string, rollbackBlocks, goroutineName string, dataTypeBlockBody int64) error {

	log.Debug("blockID", blockID)

	parser := new(Parser)
	var count int64
	blocks := make(map[int64]string)
	for {
		/*
			// note in the database that we are alive
						upd_deamon_time($db);
			// note for not to provoke cleaning of the tables
						upd_main_lock($db);
			// check if we have to get out, because the script version has been updated
						if (check_deamon_restart($db)){
							main_unlock();
							exit;
						}*/
		if blockID < 2 {
			return utils.ErrInfo(errors.New("block_id < 2"))
		}
		// if the limit of blocks received from the node was exaggerated
		var rollback = consts.RB_BLOCKS_1
		if rollbackBlocks == "rollback_blocks_2" {
			rollback = consts.RB_BLOCKS_2
		}
		if count > int64(rollback) {
			ClearTmp(blocks)
			return utils.ErrInfo(errors.New("count > variables[rollback_blocks]"))
		}

		// load the block body from the host
		binaryBlock, err := utils.GetBlockBody(host, blockID, dataTypeBlockBody)

		if err != nil {
			ClearTmp(blocks)
			return utils.ErrInfo(err)
		}
		log.Debug("binaryBlock: %x\n", binaryBlock)
		binaryBlockFull := binaryBlock
		if len(binaryBlock) == 0 {
			log.Debug("len(binaryBlock) == 0")
			ClearTmp(blocks)
			return utils.ErrInfo(errors.New("len(binaryBlock) == 0"))
		}
		converter.BytesShift(&binaryBlock, 1) // уберем 1-й байт - тип (блок/тр-я)
		// remove the 1st byte - type (block/transaction)
		// parse the heading of a block
		blockData := utils.ParseBlockHeader(&binaryBlock)
		log.Debug("blockData", blockData)

		// if the buggy chain exists, here we will ignore it
		config := &model.Config{}
		err = config.GetConfig()
		if err != nil {
			ClearTmp(blocks)
			return utils.ErrInfo(err)
		}
		badBlocks := make(map[int64]string)
		if len(config.BadBlocks) > 0 {
			err = json.Unmarshal([]byte(config.BadBlocks), &badBlocks)
			if err != nil {
				ClearTmp(blocks)
				return utils.ErrInfo(err)
			}
		}
		if badBlocks[blockData.BlockID] == string(converter.BinToHex(blockData.Sign)) {
			ClearTmp(blocks)
			return utils.ErrInfo(errors.New("bad block"))
		}
		if blockData.BlockID != blockID {
			ClearTmp(blocks)
			return utils.ErrInfo(errors.New("bad block_data['block_id']"))
		}

		// the block size cannot be more than max_block_size
		if int64(len(binaryBlock)) > syspar.GetMaxBlockSize() {
			ClearTmp(blocks)
			return utils.ErrInfo(errors.New(`len(binaryBlock) > variables.Int64["max_block_size"]`))
		}

		// we need the hash of previous block to find where the fork started
		prevBlock := &model.Block{}
		err = prevBlock.GetBlock(blockID - 1)
		if err != nil {
			ClearTmp(blocks)
			return utils.ErrInfo(err)
		}

		// we need the mrklRoot of the current block
		mrklRoot, err := utils.GetMrklroot(binaryBlock, false, syspar.GetMaxTxSize(), syspar.GetMaxTxCount())
		if err != nil {
			ClearTmp(blocks)
			return utils.ErrInfo(err)
		}

		// the public key of the one who has generated this block
		nodePublicKey, err := GetNodePublicKeyWalletOrCB(blockData.WalletID, blockData.StateID)
		if err != nil {
			return utils.ErrInfo(err)
		}

		// SIGN from 128 bytes to 512 bytes. Signature of TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT
		forSign := fmt.Sprintf("0,%v,%x,%v,%v,%v,%s", blockData.BlockID, prevBlock.Hash, blockData.Time, blockData.WalletID, blockData.StateID, mrklRoot)
		log.Debug("forSign", forSign)

		// check the signature
		_, okSignErr := utils.CheckSign([][]byte{nodePublicKey}, forSign, blockData.Sign, true)
		log.Debug("okSignErr", okSignErr)

		// save the block itself in the file, for not to load the memory
		file, err := ioutil.TempFile(*utils.Dir, "DC")
		defer os.Remove(file.Name())
		_, err = file.Write(binaryBlockFull)
		if err != nil {
			ClearTmp(blocks)
			return utils.ErrInfo(err)
		}
		blocks[blockID] = file.Name()
		blockID--
		count++

		// load the previous blocks till the hash of previous one is different
		// in other words, while the signature with prevBlockHash is incorrect, so far there is something in okSignErr
		if okSignErr == nil {
			log.Debug("plug found blockID=%v\n", blockData.BlockID)
			break
		}
	}

	// to take the blocks in order
	blocksSorted := converter.SortMap(blocks)
	log.Debug("blocks", blocksSorted)

	logging.WriteSelectiveLog("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")

	affect, err := model.MarkVerifiedAndNotUsedTransactionsUnverified()
	if err != nil {
		logging.WriteSelectiveLog(err)
		return utils.ErrInfo(err)
	}
	logging.WriteSelectiveLog("affect: " + converter.Int64ToStr(affect))

	// we roll back our blocks before fork started
	block := &model.Block{}
	myRollbackBlocks, err := block.GetBlocksFrom(blockID, "desc")
	if err != nil {
		return p.ErrInfo(err)
	}
	for _, block := range myRollbackBlocks {
		log.Debug("We roll away blocks before plug", blockID)
		parser.GoroutineName = goroutineName
		parser.BinaryData = block.Data
		err = parser.ParseDataRollback()
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	log.Debug("blocks", blocksSorted)

	prevBlock := make(map[int64]*utils.BlockData)

	// go through the new blocks
	for _, data := range blocksSorted {
		for intBlockID, tmpFileName := range data {
			log.Debug("Go on new blocks", intBlockID, tmpFileName)

			// check and record the data
			binaryBlock, err := ioutil.ReadFile(tmpFileName)
			if err != nil {
				return utils.ErrInfo(err)
			}
			log.Debug("binaryBlock: %x\n", binaryBlock)
			parser.GoroutineName = goroutineName
			parser.BinaryData = binaryBlock
			// we pass the information about the previous block. So far there are new blocks, information about previous blocks in blockchain is still old, because the updating of blockchain is going below
			if prevBlock[intBlockID-1] != nil {
				log.Debug("prevBlock[intBlockID-1] != nil : %v", prevBlock[intBlockID-1])
				parser.PrevBlock.Hash = prevBlock[intBlockID-1].Hash
				parser.PrevBlock.Time = prevBlock[intBlockID-1].Time
				parser.PrevBlock.BlockID = prevBlock[intBlockID-1].BlockID
			}

			// If the error returned, then the transferred block has already rolled back
			// info_block и config.my_block_id are uploading only when there is no error
			err0 := parser.ParseDataFull(false)
			// we will get hashes and time for the further processing
			if err0 == nil {
				prevBlock[intBlockID] = parser.GetBlockInfo()
				log.Debug("prevBlock[%d] = %v", intBlockID, prevBlock[intBlockID])
			}
			// if the mistake happened, we roll back all previous blocks from new chain
			if err0 != nil {
				parser.BlockError(err) // why?
				log.Debug("there is an error is rolled back all previous blocks of a new chain: %v", err)

				// we ban the host which gave us a false chain for 1 hour
				// necessarily go through the blocks in reverse order
				blocksSorted := converter.RSortMap(blocks)
				for _, data := range blocksSorted {
					for int2BlockID, tmpFileName := range data {
						log.Debug("int2BlockID", int2BlockID)
						if int2BlockID >= intBlockID {
							continue
						}
						binaryBlock, err := ioutil.ReadFile(tmpFileName)
						if err != nil {
							return utils.ErrInfo(err)
						}
						parser.GoroutineName = goroutineName
						parser.BinaryData = binaryBlock
						err = parser.ParseDataRollback()
						if err != nil {
							return utils.ErrInfo(err)
						}
					}
				}
				// we insert from block_chain our data which was before
				log.Debug("We push data from our block_chain, which were previously")
				block := &model.Block{}
				beforeBlocks, err := block.GetBlocksFrom(blockID, "asc")
				if err != nil {
					return p.ErrInfo(err)
				}
				for _, block := range beforeBlocks {
					log.Debug("blockID", blockID, "intBlockID", intBlockID)
					parser.GoroutineName = goroutineName
					parser.BinaryData = block.Data
					err = parser.ParseDataFull(false)
					if err != nil {
						return utils.ErrInfo(err)
					}
				}
				// because in the previous request to block_chain the data could be absent, because the $block_id is bigger than our the biggest id in block_chain
				// that means the info_block could not be updated and could stay away from adding new blocks, which will result in skipping the block in block_chain
				lastMyBlock := &model.Block{}
				err = lastMyBlock.GetMaxBlock()
				if err != nil {
					return utils.ErrInfo(err)
				}
				binary := lastMyBlock.Data
				converter.BytesShift(&binary, 1) // remove the first byte which is the type (block/territory)
				lastMyBlockData := utils.ParseBlockHeader(&binary)
				infoBlock := &model.InfoBlock{
					Hash:    lastMyBlock.Hash,
					BlockID: lastMyBlockData.BlockID,
					Time:    lastMyBlockData.Time,
					Sent:    0,
				}
				err = infoBlock.Update()
				if err != nil {
					return utils.ErrInfo(err)
				}
				err = model.UpdateConfig("my_block_id", converter.Int64ToStr(lastMyBlockData.BlockID))
				if err != nil {
					return utils.ErrInfo(err)
				}
				ClearTmp(blocks)
				return utils.ErrInfo(err0) // go to the next block in queue_blocks
			}
		}
	}
	log.Debug("remove the blocks and enter new block_chain")

	// if all was recorded without errors, delete the blocks from block_chain and insert new
	blockFrom := &model.Block{ID: blockID}
	err = blockFrom.DeleteChain()
	if err != nil {
		return utils.ErrInfo(err)
	}
	log.Debug("prevblock", prevBlock)
	log.Debug("blocks", blocks)

	// go through new blocks
	bSorted := converter.SortMap(blocks)
	log.Debug("blocksSorted_", bSorted)
	for _, data := range bSorted {
		for blockID, tmpFileName := range data {

			block, err := ioutil.ReadFile(tmpFileName)
			if err != nil {
				return utils.ErrInfo(err)
			}
			blockHex := converter.BinToHex(block)

			// record in the chain of blocks
			b := prevBlock[blockID]
			ibUpdate := &model.InfoBlock{Hash: b.Hash, BlockID: b.BlockID, Time: b.Time, WalletID: b.WalletID, StateID: b.StateID}
			err = ibUpdate.Update()
			if err != nil {
				return utils.ErrInfo(err)
			}
			err = model.UpdateConfig("my_block_id", converter.Int64ToStr(b.BlockID))
			if err != nil {
				return utils.ErrInfo(err)
			}

			// because this data we made by ourselves, so you can record them directly to the table of verified data, that will be send to other nodes
			existsB := &model.Block{}
			exists, err := existsB.IsExistsID(blockID)
			if err != nil {
				return utils.ErrInfo(err)
			}
			if !exists {
				b := prevBlock[blockID]
				block := &model.Block{
					ID:       blockID,
					Hash:     b.Hash,
					StateID:  b.StateID,
					WalletID: b.WalletID,
					Time:     b.Time,
					Data:     blockHex,
				}
				err := block.Create()
				if err != nil {
					return utils.ErrInfo(err)
				}
			}
			err = os.Remove(tmpFileName)
			if err != nil {
				return utils.ErrInfo(err)
			}
			log.Debug("tmpFileName %v", tmpFileName)
		}
	}
	log.Debug("HAPPY END")
	return nil
}
