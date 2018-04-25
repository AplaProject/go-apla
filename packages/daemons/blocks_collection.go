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
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/parser"
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
	hosts := syspar.GetRemoteHosts()
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
				banNode(host, err)
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
				err := parser.GetBlocks(block.Header.BlockID-1, host)
				if err != nil {
					d.logger.WithFields(log.Fields{"error": err, "type": consts.ParserError}).Error("processing block")
					banNode(host, err)
					return err
				}
			}

			block.PrevHeader, err = parser.GetBlockDataFromBlockChain(block.Header.BlockID - 1)
			if err != nil {
				banNode(host, err)
				return utils.ErrInfo(fmt.Errorf("can't get block %d", block.Header.BlockID-1))
			}
			if err = block.CheckBlock(); err != nil {
				banNode(host, err)
				return err
			}
			if err = block.PlayBlockSafe(); err != nil {
				banNode(host, err)
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

func banNode(host string, err error) {
	// TODO
}
