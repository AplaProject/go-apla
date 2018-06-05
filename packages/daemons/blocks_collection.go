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

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/rollback"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/tcpserver"
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

	infoBlock := &model.InfoBlock{}
	found, err := infoBlock.Get()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting cur blockID")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Error("Info block not found")
		return errors.New("Info block not found")
	}

	if infoBlock.BlockID >= maxBlockID {
		log.WithFields(log.Fields{"blockID": infoBlock.BlockID, "maxBlockID": maxBlockID}).Debug("Max block is already in the host")
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
	curBlock := &model.InfoBlock{}
	if _, err := curBlock.Get(); err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting info block")
		return err
	}

	if ctx.Err() != nil {
		d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
		return ctx.Err()
	}

	playRawBlock := func(rawBlocksQueueCh chan []byte) error {
		for rb := range rawBlocksQueueCh {
			block, err := parser.ProcessBlockWherePrevFromBlockchainTable(rb, true)
			if err != nil {
				// we got bad block and should ban this host
				banNode(host, block, err)
				d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("processing block")
				return err
			}

			// hash compare could be failed in the case of fork
			hashMatched, thisErrIsOk := block.CheckHash()
			if thisErrIsOk != nil {
				d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("checking block hash")
			}

			if !hashMatched {
				//it should be fork, replace our previous blocks to ones from the host
				err := GetBlocks(block.Header.BlockID-1, host)
				if err != nil {
					d.logger.WithFields(log.Fields{"error": err, "type": consts.ParserError}).Error("processing block")
					banNode(host, block, err)
					return err
				}
			}

			block.PrevHeader, err = parser.GetBlockDataFromBlockChain(block.Header.BlockID - 1)
			if err != nil {
				banNode(host, block, err)
				return utils.ErrInfo(fmt.Errorf("can't get block %d", block.Header.BlockID-1))
			}
			if err = block.CheckBlock(); err != nil {
				banNode(host, block, err)
				return err
			}
			if err = block.PlayBlockSafe(); err != nil {
				banNode(host, block, err)
				return err
			}
		}
		return nil
	}

	st := time.Now()
	d.logger.Infof("starting downloading blocks from %d to %d (%d) \n", curBlock.BlockID, maxBlockID, maxBlockID-curBlock.BlockID)

	count := 0
	var err error
	for blockID := curBlock.BlockID + 1; blockID <= maxBlockID; blockID += int64(tcpserver.BlocksPerRequest) {
		var rawBlocksChan chan []byte
		rawBlocksChan, err = utils.GetBlocksBody(host, blockID, tcpserver.BlocksPerRequest, consts.DATA_TYPE_BLOCK_BODY, false)
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

	if err = parser.InsertBlockWOForks(newBlock, false, true); err != nil {
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
	infoBlock := &model.InfoBlock{}
	_, err := infoBlock.Get()
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting info block")
		return false, err
	}
	// we have empty blockchain, we need to load blockchain from file or other source
	if infoBlock.BlockID == 0 {
		logger.Debug("blockchain should be loaded")
		return true, nil
	}
	return false, nil
}

func banNode(host string, block *parser.Block, err error) {
	var (
		reason             string
		blockId, blockTime int64
	)
	if err != nil {
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
func GetBlocks(blockID int64, host string) error {
	blocks, err := getBlocks(blockID, host)
	if err != nil {
		return err
	}

	// mark all transaction as unverified
	_, err = model.MarkVerifiedAndNotUsedTransactionsUnverified()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"type":  consts.DBError,
		}).Error("marking verified and not used transactions unverified")
		return utils.ErrInfo(err)
	}

	// get starting blockID from slice of blocks
	if len(blocks) > 0 {
		blockID = blocks[len(blocks)-1].Header.BlockID
	}

	// we have the slice of blocks for applying
	// first of all we should rollback old blocks
	block := &model.Block{}
	myRollbackBlocks, err := block.GetBlocksFrom(blockID-1, "desc", 0)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting rollback blocks from blockID")
		return utils.ErrInfo(err)
	}
	for _, block := range myRollbackBlocks {
		err := rollback.RollbackBlock(block.Data, false)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	return processBlocks(blocks)
}

func getBlocks(blockID int64, host string) ([]*parser.Block, error) {
	rollback := syspar.GetRbBlocks1()

	badBlocks := make(map[int64]string)

	blocks := make([]*parser.Block, 0)
	var count int64

	// load the block bodies from the host
	blocksCh, err := utils.GetBlocksBody(host, blockID, tcpserver.BlocksPerRequest, consts.DATA_TYPE_BLOCK_BODY, true)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	for binaryBlock := range blocksCh {
		if blockID < 2 {
			break
		}

		// if the limit of blocks received from the node was exaggerated
		if count > int64(rollback) {
			break
		}

		block, err := parser.ProcessBlockWherePrevFromBlockchainTable(binaryBlock, true)
		if err != nil {
			return nil, utils.ErrInfo(err)
		}

		if badBlocks[block.Header.BlockID] == string(converter.BinToHex(block.Header.Sign)) {
			log.WithFields(log.Fields{"block_id": block.Header.BlockID, "type": consts.InvalidObject}).Error("block is bad")
			return nil, utils.ErrInfo(errors.New("bad block"))
		}
		if block.Header.BlockID != blockID {
			log.WithFields(log.Fields{"header_block_id": block.Header.BlockID, "block_id": blockID, "type": consts.InvalidObject}).Error("block ids does not match")
			return nil, utils.ErrInfo(errors.New("bad block_data['block_id']"))
		}

		// TODO: add checking for MAX_BLOCK_SIZE

		// the public key of the one who has generated this block
		nodePublicKey, err := syspar.GetNodePublicKeyByPosition(block.Header.NodePosition)
		if err != nil {
			log.WithFields(log.Fields{"header_block_id": block.Header.BlockID, "block_id": blockID, "type": consts.InvalidObject}).Error("block ids does not match")
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
		blockID--
		count++

		// check the signature
		_, okSignErr := utils.CheckSign([][]byte{nodePublicKey}, forSign, block.Header.Sign, true)
		if okSignErr == nil {
			break
		}
	}

	return blocks, nil
}

func processBlocks(blocks []*parser.Block) error {
	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("starting transaction")
		return utils.ErrInfo(err)
	}

	// go through new blocks from the smallest block_id to the largest block_id
	prevBlocks := make(map[int64]*parser.Block, 0)

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

		if err := block.PlayBlock(dbTransaction); err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}
		prevBlocks[block.Header.BlockID] = block

		// for last block we should update block info
		if i == 0 {
			err := parser.UpdBlockInfo(dbTransaction, block)
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
		if err := parser.InsertIntoBlockchain(dbTransaction, block); err != nil {
			dbTransaction.Rollback()
			return err
		}
	}

	return dbTransaction.Commit()
}
