// MIT License
//
// Copyright (c) 2016 GenesisCommunity
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
package daemons

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/AplaProject/go-apla/packages/rollback"
	"github.com/GenesisCommunity/go-genesis/packages/block"
	"github.com/GenesisCommunity/go-genesis/packages/conf"
	"github.com/GenesisCommunity/go-genesis/packages/conf/syspar"
	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/crypto"
	"github.com/GenesisCommunity/go-genesis/packages/model"
	"github.com/GenesisCommunity/go-genesis/packages/service"
	"github.com/GenesisCommunity/go-genesis/packages/tcpserver"
	"github.com/GenesisCommunity/go-genesis/packages/transaction"
	"github.com/GenesisCommunity/go-genesis/packages/utils"

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
				err := GetBlocks(b.Header.BlockID-1, host)
				if err != nil {
					d.logger.WithFields(log.Fields{"error": err, "type": consts.ParserError}).Error("processing block")
					banNode(host, b, err)
					return err
				}
			}

			b.PrevHeader, err = block.GetBlockDataFromBlockChain(b.Header.BlockID - 1)
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

func banNode(host string, block *block.Block, err error) {
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
func GetBlocks(blockID int64, host string) error {
	blocks, err := getBlocks(blockID, host)
	if err != nil {
		return err
	}
	transaction.CleanCache()

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

func getBlocks(blockID int64, host string) ([]*block.Block, error) {
	rollback := syspar.GetRbBlocks1()

	badBlocks := make(map[int64]string)

	blocks := make([]*block.Block, 0)
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

		block, err := block.ProcessBlockWherePrevFromBlockchainTable(binaryBlock, true)
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

func processBlocks(blocks []*block.Block) error {
	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("starting transaction")
		return utils.ErrInfo(err)
	}

	// go through new blocks from the smallest block_id to the largest block_id
	prevBlocks := make(map[int64]*block.Block, 0)

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
			return err
		}

		if err := b.Play(dbTransaction); err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}
		prevBlocks[b.Header.BlockID] = b

		// for last block we should update block info
		if i == 0 {
			err := block.UpdBlockInfo(dbTransaction, b)
			if err != nil {
				dbTransaction.Rollback()
				return utils.ErrInfo(err)
			}
		}
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
		bl := &model.Block{}
		err = bl.DeleteById(dbTransaction, b.Header.BlockID)
		if err != nil {
			dbTransaction.Rollback()
			return err
		}
		// insert new blocks into blockchain
		if err := block.InsertIntoBlockchain(dbTransaction, b); err != nil {
			dbTransaction.Rollback()
			return err
		}
	}

	return dbTransaction.Commit()
}
