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
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/block"
	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/rollback"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/tcpserver"
	"github.com/GenesisKernel/go-genesis/packages/transaction"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// ErrNodesUnavailable is returned when all nodes is unavailable
var ErrNodesUnavailable = errors.New("All nodes unavailable")

// BlocksCollection collects and parses blocks
func BlocksCollection(ctx context.Context, d *daemon) error {
	if ctx.Err() != nil {
		d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
		return ctx.Err()
	}

	return blocksCollection(ctx, d)
}

func InitialLoad(logger *log.Entry) error {

	// check for initial load
	toLoad, err := needLoad(logger)
	if err != nil {
		return err
	}

	if toLoad {
		logger.Debug("start first block loading")

		if err := firstLoad(logger); err != nil {
			return err
		}
	}

	return nil
}

func blocksCollection(ctx context.Context, d *daemon) (err error) {
	hosts, err := filterBannedHosts(syspar.GetRemoteHosts())
	if err != nil {
		return err
	}
	var (
		chooseFromConfig bool
		host             string
		maxBlockID       int64
	)
	if len(hosts) > 0 {
		// get a host with the biggest block id from system parameters
		host, maxBlockID, err = utils.ChooseBestHost(ctx, hosts, d.logger)
		if err != nil {
			if err == utils.ErrNodesUnavailable {
				chooseFromConfig = true
			} else {
				return err
			}
		}
	} else {
		chooseFromConfig = true
	}

	if chooseFromConfig {
		// get a host with the biggest block id from config
		log.Debug("Getting a host with biggest block from config")
		hosts = conf.GetNodesAddr()
		if len(hosts) > 0 {
			host, maxBlockID, err = utils.ChooseBestHost(ctx, hosts, d.logger)
			if err != nil {
				return err
			}
		}
	}

	lastBlock, found, err := blockchain.GetLastBlock()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("Getting last block")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Error("last block not found")
		return errors.New("Info block not found")
	}

	if lastBlock.Header.BlockID >= maxBlockID {
		log.WithFields(log.Fields{"blockID": lastBlock.Header.BlockID, "maxBlockID": maxBlockID}).Debug("Max block is already in the host")
		return nil
	}

	DBLock()
	defer func() {
		DBUnlock()
		service.NodeDoneUpdatingBlockchain()
	}()

	// update our chain till maxBlockID from the host
	return UpdateChain(ctx, d, host, maxBlockID)
}

// UpdateChain load from host all blocks from our last block to maxBlockID
func UpdateChain(ctx context.Context, d *daemon, host string, maxBlockID int64) error {

	// get current block id from our blockchain
	curBlock, found, err := blockchain.GetLastBlock()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting last block")
		return err
	}

	if ctx.Err() != nil {
		d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
		return ctx.Err()
	}

	playRawBlock := func(rawBlocksQueueCh chan []byte) error {
		for rb := range rawBlocksQueueCh {
			b, err := block.ProcessBlockWherePrevFromBlockchainTable(rb, true)
			if err != nil {
				// we got bad block and should ban this host
				banNode(host, b, err)
				d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("processing block")
				return err
			}

			// hash compare could be failed in the case of fork
			hashMatched, thisErrIsOk := b.CheckHash()
			if thisErrIsOk != nil {
				d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("checking block hash")
			}

			if !hashMatched {
				transaction.CleanCache()
				//it should be fork, replace our previous blocks to ones from the host
				err := GetBlocks(b.PrevHeader.Hash, host)
				if err != nil {
					d.logger.WithFields(log.Fields{"error": err, "type": consts.ParserError}).Error("processing block")
					banNode(host, b, err)
					return err
				}
			}

			b.PrevHeader, err = block.GetBlockDataFromBlockChain(b.PrevHeader.Hash)
			if err != nil {
				banNode(host, b, err)
				return utils.ErrInfo(fmt.Errorf("can't get block %d", b.Header.BlockID-1))
			}
			if err = b.Check(); err != nil {
				banNode(host, b, err)
				return err
			}
			if err = b.PlaySafe(); err != nil {
				banNode(host, b, err)
				return err
			}
		}
		return nil
	}

	st := time.Now()
	d.logger.Infof("starting downloading blocks from %d to %d (%d) \n", curBlock.Header.BlockID, maxBlockID, maxBlockID-curBlock.Header.BlockID)

	count := 0
	blockHash := curBlock.NextHash
	curBlockID := curBlock.Header.BlockID + 1
	nextBlock, found, err := blockchain.GetBlock(curBlock.NextHash)
	if err != nil {
		return err
	}
	if !found {
		blockHash = curBlock.Header.Hash
		curBlockID = curBlock.Header.BlockID
	}
	for blockID := curBlockID; blockID <= maxBlockID; blockID += int64(tcpserver.BlocksPerRequest) {
		var rawBlocksChan chan []byte
		rawBlocksChan, err = utils.GetBlocksBody(host, blockHash, tcpserver.BlocksPerRequest, consts.DATA_TYPE_BLOCK_BODY, false)
		if err != nil {
			d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("getting block body")
			break
		}

		err = playRawBlock(rawBlocksChan)
		if err != nil {
			d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("playing raw block")
			break
		}
		count++
		blocks, err := blockchain.GetNBlocksFrom(nextBlock.Header.Hash, int(tcpserver.BlocksPerRequest), 1)
		if err != nil {
			return err
		}
		blockHash = blocks[len(blocks)-1].Header.Hash
		curBlockID = blocks[len(blocks)-1].Header.BlockID
	}

	if err != nil {
		d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("retrieving blockchain from node")
	} else {
		d.logger.Infof("%d blocks was collected (%s) \n", count, time.Since(st).String())
	}
	return err
}

// init first block from file or from embedded value
func loadFirstBlock(logger *log.Entry) error {
	newBlock, err := ioutil.ReadFile(conf.Config.FirstBlockPath)
	if err != nil {
		logger.WithFields(log.Fields{
			"type": consts.IOError, "error": err, "path": conf.Config.FirstBlockPath,
		}).Error("reading first block from file")
	}

	if err = block.InsertBlockWOForks(newBlock, false, true); err != nil {
		logger.WithFields(log.Fields{"type": consts.ParserError, "error": err}).Error("inserting new block")
		return err
	}

	return nil
}

func firstLoad(logger *log.Entry) error {
	DBLock()
	defer DBUnlock()

	return loadFirstBlock(logger)
}

func needLoad(logger *log.Entry) (bool, error) {
	_, found, err := blockchain.GetLastBlock()
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting info block")
		return false, err
	}
	// we have empty blockchain, we need to load blockchain from file or other source
	if !found {
		logger.Debug("blockchain should be loaded")
		return true, nil
	}
	return false, nil
}

func banNode(host string, block *block.PlayableBlock, err error) {
	var (
		reason             string
		blockId, blockTime int64
	)
	if err != nil {
		if err == transaction.ErrDuplicatedTx {
			return
		}
		reason = err.Error()
	}

	if block != nil {
		blockId, blockTime = block.Header.BlockID, block.Header.Time
	} else {
		blockId, blockTime = -1, time.Now().Unix()
	}

	log.WithFields(log.Fields{"reason": reason, "host": host, "block_id": blockId, "block_time": blockTime}).Debug("ban node")

	n, err := syspar.GetNodeByHost(host)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("getting node by host")
		return
	}

	err = service.GetNodesBanService().RegisterBadBlock(n, blockId, blockTime, reason)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "node": n.KeyID, "block": blockId}).Error("registering bad block from node")
	}
}

func filterBannedHosts(hosts []string) ([]string, error) {
	var goodHosts []string
	for _, h := range hosts {
		n, err := syspar.GetNodeByHost(h)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("getting node by host")
			return nil, err
		}

		if !service.GetNodesBanService().IsBanned(n) {
			goodHosts = append(goodHosts, n.TCPAddress)
		}
	}
	return goodHosts, nil
}

// GetBlocks is returning blocks
func GetBlocks(blockHash []byte, host string) error {
	blocks, err := getBlocks(blockHash, host)
	if err != nil {
		return err
	}

	// get starting blockID from slice of blocks
	if len(blocks) > 0 {
		blockHash = blocks[len(blocks)-1].Header.Hash
	}

	// we have the slice of blocks for applying
	// first of all we should rollback old blocks
	myRollbackBlocks, err := blockchain.GetNBlocksFrom(blockHash, -1, 0)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting rollback blocks from blockID")
		return utils.ErrInfo(err)
	}
	for _, block := range myRollbackBlocks {
		err := rollback.RollbackBlock(block, false)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	return processBlocks(blocks)
}

func getBlocks(blockHash []byte, host string) ([]*block.PlayableBlock, error) {
	rollback := syspar.GetRbBlocks1()

	blocks := make([]*block.PlayableBlock, 0)
	var count int64

	// load the block bodies from the host
	blocksCh, err := utils.GetBlocksBody(host, blockHash, tcpserver.BlocksPerRequest, consts.DATA_TYPE_BLOCK_BODY, true)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	for binaryBlock := range blocksCh {
		// if the limit of blocks received from the node was exaggerated
		if count > int64(rollback) {
			break
		}

		block, err := block.ProcessBlockWherePrevFromBlockchainTable(binaryBlock, true)
		if err != nil {
			return nil, utils.ErrInfo(err)
		}

		if string(block.Header.Hash) != string(blockHash) {
			log.WithFields(log.Fields{"header_block_hash": block.Header.Hash, "block_id": blockHash, "type": consts.InvalidObject}).Error("block hashes does not match")
			return nil, utils.ErrInfo(errors.New("bad block_data['block_id']"))
		}

		// TODO: add checking for MAX_BLOCK_SIZE

		// the public key of the one who has generated this block
		nodePublicKey, err := syspar.GetNodePublicKeyByPosition(block.Header.NodePosition)
		if err != nil {
			log.WithFields(log.Fields{"header_block_hash": block.Header.Hash, "block_id": blockHash, "type": consts.InvalidObject}).Error("block ids does not match")
			return nil, utils.ErrInfo(err)
		}

		// SIGN from 128 bytes to 512 bytes. Signature of TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT
		forSign := fmt.Sprintf("0,%v,%x,%v,%v,%v,%v,%s",
			block.Header.BlockID, block.PrevHeader.Hash, block.Header.Time,
			block.Header.EcosystemID, block.Header.KeyID, block.Header.NodePosition,
			block.MrklRoot,
		)

		// save the block
		blocks = append(blocks, block)
		count++

		// check the signature
		_, okSignErr := utils.CheckSign([][]byte{nodePublicKey}, forSign, block.Header.Sign, true)
		if okSignErr == nil {
			break
		}
	}

	return blocks, nil
}

func processBlocks(blocks []*block.PlayableBlock) error {
	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("starting transaction")
		return utils.ErrInfo(err)
	}

	// go through new blocks from the smallest block_id to the largest block_id
	prevBlocks := make(map[int64]*block.PlayableBlock, 0)

	for i := len(blocks) - 1; i >= 0; i-- {
		b := blocks[i]

		if prevBlocks[b.Header.BlockID-1] != nil {
			b.PrevHeader.Hash = prevBlocks[b.Header.BlockID-1].Header.Hash
			b.PrevHeader.Time = prevBlocks[b.Header.BlockID-1].Header.Time
			b.PrevHeader.BlockID = prevBlocks[b.Header.BlockID-1].Header.BlockID
			b.PrevHeader.EcosystemID = prevBlocks[b.Header.BlockID-1].Header.EcosystemID
			b.PrevHeader.KeyID = prevBlocks[b.Header.BlockID-1].Header.KeyID
			b.PrevHeader.NodePosition = prevBlocks[b.Header.BlockID-1].Header.NodePosition
		}

		forSha := fmt.Sprintf("%d,%x,%s,%d,%d,%d,%d", b.Header.BlockID, b.PrevHeader.Hash, b.MrklRoot, b.Header.Time, b.Header.EcosystemID, b.Header.KeyID, b.Header.NodePosition)
		hash, err := crypto.DoubleHash([]byte(forSha))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("double hashing block")
		}
		b.Header.Hash = hash

		if err := b.Check(); err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}

		if err := b.Play(dbTransaction); err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}
		prevBlocks[b.Header.BlockID] = b

		if b.SysUpdate {
			if err := syspar.SysUpdate(dbTransaction); err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
				return utils.ErrInfo(err)
			}
		}
	}

	// If all right we can delete old blockchain and write new
	for i := len(blocks) - 1; i >= 0; i-- {
		b := blocks[i]
		// Delete old blocks from blockchain
		err = blockchain.DeleteBlock(b.Header.Hash)
		if err != nil {
			dbTransaction.Rollback()
			return err
		}
		// insert new blocks into blockchain
		nodePrivKey, _, err := utils.GetNodeKeys()
		if err != nil {
			return err
		}
		if err := blockchain.InsertBlock(b.Header.Hash, b.ToBlockchainBlock(), nodePrivKey); err != nil {
			dbTransaction.Rollback()
			return err
		}
	}

	return dbTransaction.Commit()
}
