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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context/ctxhttp"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/tcpserver"
	"github.com/GenesisKernel/go-genesis/packages/utils"
)

// BlocksCollection collects and parses blocks
func BlocksCollection(ctx context.Context, d *daemon) error {
	if err := initialLoad(ctx, d); err != nil {
		return err
	}

	if ctx.Err() != nil {
		d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
		return ctx.Err()
	}

	return blocksCollection(ctx, d)
}

func initialLoad(ctx context.Context, d *daemon) error {

	// check for initial load
	toLoad, err := needLoad(d.logger)
	if err != nil {
		return err
	}

	if toLoad {
		d.logger.Debug("start first block loading")

		if err := firstLoad(ctx, d); err != nil {
			return err
		}
	}

	return nil
}

func blocksCollection(ctx context.Context, d *daemon) error {

	hosts := syspar.GetRemoteHosts()

	// get a host with the biggest block id
	host, maxBlockID, err := utils.ChooseBestHost(ctx, hosts, d.logger)
	if err != nil {
		return err
	}

	// NOTE: should be generalized in separate method
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
	defer DBUnlock()
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
			block, err := parser.ProcessBlockWherePrevFromBlockchainTable(rb)
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

func downloadChain(ctx context.Context, fileName, url string, logger *log.Entry) error {

	for i := 0; i < consts.DOWNLOAD_CHAIN_TRY_COUNT; i++ {
		loadCtx, cancel := context.WithTimeout(ctx, 30)
		defer cancel()

		_, err := downloadToFile(loadCtx, url, fileName, logger)
		if err != nil {
			continue
		}
	}
	return fmt.Errorf("can't download blockchain from %s", url)
}

// init first block from file or from embedded value
func loadFirstBlock(logger *log.Entry) error {

	newBlock, err := ioutil.ReadFile(*conf.FirstBlockPath)
	if err != nil {
		logger.WithFields(log.Fields{
			"type": consts.IOError, "error": err, "path": *conf.FirstBlockPath,
		}).Error("reading first block from file")
	}

	if err = parser.InsertBlockWOForks(newBlock); err != nil {
		logger.WithFields(log.Fields{"type": consts.ParserError, "error": err}).Error("inserting new block")
		return err
	}

	return nil
}

func firstLoad(ctx context.Context, d *daemon) error {

	DBLock()
	defer DBUnlock()

	return loadFirstBlock(d.logger)
}

func needLoad(logger *log.Entry) (bool, error) {
	infoBlock := &model.InfoBlock{}
	_, err := infoBlock.Get()
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting info block")
		return false, err
	}
	// we have empty blockchain, we need to load blockchain from file or other source
	if infoBlock.BlockID == 0 || *conf.StartBlockID > 0 {
		logger.Debug("blockchain should be loaded")
		return true, nil
	}
	return false, nil
}

func banNode(host string, err error) {
	// TODO
}

func loadFromFile(ctx context.Context, fileName string, logger *log.Entry) error {
	file, err := os.Open(fileName)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("opening file, to load blockhain from it")
		return err
	}
	defer file.Close()
	for {
		if ctx.Err() != nil {
			logger.WithFields(log.Fields{"type": consts.ContextError, "error": err}).Error("context error")
			return ctx.Err()
		}

		block, err := readBlock(file, logger)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if block == nil {
			return nil
		}

		if *conf.EndBlockID > 0 && block.ID == *conf.EndBlockID {
			return nil
		}

		if *conf.StartBlockID == 0 || (*conf.StartBlockID > 0 && block.ID > *conf.StartBlockID) {
			if err = parser.InsertBlockWOForks(block.Data); err != nil {
				return err
			}
		}
	}
}

// downloadToFile downloads and saves the specified file
func downloadToFile(ctx context.Context, url, file string, logger *log.Entry) (int64, error) {
	resp, err := ctxhttp.Get(ctx, &http.Client{}, url)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ContextError, "error": err, "url": url}).Error("context error")
		return 0, utils.ErrInfo(err)
	}
	defer resp.Body.Close()

	f, err := os.Create(file)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("creating file for writing downloaded blockchain")
		return 0, utils.ErrInfo(err)
	}
	defer f.Close()

	var offset int64
	for {
		if ctx.Err() != nil {
			logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
			return 0, ctx.Err()
		}

		data, err := ioutil.ReadAll(io.LimitReader(resp.Body, 10000))
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "url": url}).Error("downloading file from url")
			return offset, utils.ErrInfo(err)
		}

		f.WriteAt(data, offset)
		offset += int64(len(data))
		if len(data) == 0 {
			break
		}
	}
	return offset, nil
}
