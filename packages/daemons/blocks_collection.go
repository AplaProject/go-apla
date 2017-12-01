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
	"sync"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context/ctxhttp"
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
	host, maxBlockID, err := chooseBestHost(ctx, hosts, d.logger)
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
	return UpdateChain(ctx, d, host, maxBlockID, "rollback_blocks_2")
}

// best host is a host with the biggest last block ID
func chooseBestHost(ctx context.Context, hosts []string, logger *log.Entry) (string, int64, error) {
	type blockAndHost struct {
		host    string
		blockID int64
		err     error
	}
	c := make(chan blockAndHost, len(hosts))

	var wg sync.WaitGroup
	for _, h := range hosts {
		if ctx.Err() != nil {
			logger.WithFields(log.Fields{"error": ctx.Err(), "type": consts.ContextError}).Error("context error")
			return "", 0, ctx.Err()
		}
		wg.Add(1)

		go func(host string) {
			blockID, err := getHostBlockID(host, logger)
			wg.Done()

			c <- blockAndHost{
				host:    host,
				blockID: blockID,
				err:     err,
			}
		}(getHostPort(h))
	}
	wg.Wait()

	maxBlockID := int64(-1)
	var bestHost string
	for i := 0; i < len(hosts); i++ {
		bl := <-c

		if bl.blockID > maxBlockID {
			maxBlockID = bl.blockID
			bestHost = bl.host
		}
	}

	return bestHost, maxBlockID, nil
}

func getHostBlockID(host string, logger *log.Entry) (int64, error) {
	conn, err := utils.TCPConn(host)
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.ConnectionError, "host": host}).Debug("error connecting to host")
		return 0, err
	}
	defer conn.Close()

	// get max block request
	_, err = conn.Write(converter.DecToBin(consts.DATA_TYPE_MAX_BLOCK_ID, 2))
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.ConnectionError, "host": host}).Error("writing max block id to host")
		return 0, err
	}

	// response
	blockIDBin := make([]byte, 4)
	_, err = conn.Read(blockIDBin)
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.ConnectionError, "host": host}).Error("reading max block id from host")
		return 0, err
	}

	return converter.BinToDec(blockIDBin), nil
}

// UpdateChain load from host all blocks from our last block to maxBlockID
func UpdateChain(ctx context.Context, d *daemon, host string, maxBlockID int64, rollbackBlocks string) error {

	// get current block id from our blockchain
	curBlock := &model.InfoBlock{}
	if _, err := curBlock.Get(); err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting info block")
		return err
	}

	for blockID := curBlock.BlockID + 1; blockID <= maxBlockID; blockID++ {
		if ctx.Err() != nil {
			d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
			return ctx.Err()
		}

		blockBin, err := utils.GetBlockBody(host, blockID, consts.DATA_TYPE_BLOCK_BODY)
		if err != nil {
			d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("getting block body")
			return err
		}

		block, err := parser.ProcessBlockWherePrevFromBlockchainTable(blockBin)
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
			// it should be fork, replace our previous blocks to ones from the host
			err := parser.GetBlocks(blockID-1, host, rollbackBlocks)
			if err != nil {
				d.logger.WithFields(log.Fields{"error": err, "type": consts.ParserError}).Error("processing block")
				banNode(host, err)
				return err
			}
		} else {
			/* TODO should we uncomment this ?????????????
			_, err := model.MarkTransactionsUnverified()
			if err != nil {
				return err
			}
			*/
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
	var newBlock []byte
	var err error

	// if len(*utils.FirstBlockDir) > 0 {
	// 	fileName := *utils.FirstBlockDir + "/1block"
	// 	logger.WithFields(log.Fields{"file_name": fileName}).Info("loading first block from file")
	// 	newBlock, err = ioutil.ReadFile(fileName)
	// 	if err != nil {
	// 		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "file_name": fileName}).Error("reading first block from file")
	// 	}
	// } else {
	// 	logger.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("FirstBlockDir doesn't set")
	// }

	newBlock, err = ioutil.ReadFile(conf.Config.FirstBlockPath)
	if err != nil {
		logger.WithFields(log.Fields{
			"type": consts.IOError, "error": err, "file_name": conf.Config.FirstBlockPath,
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

	// if conf.Config.FirstLoadBlockchain == "file" {
	// 	blockchainURL := conf.Config.FirstLoadBlockchainURL
	// 	if len(blockchainURL) == 0 {
	// 		blockchainURL = syspar.GetBlockchainURL()
	// 	}

	// 	fileName := conf.Config.WorkDir + "/public/blockchain" // !!! filepath.Join()
	// 	if err := downloadChain(ctx, fileName, blockchainURL, d.logger); err != nil {
	// 		return err
	// 	}

	// 	return loadFromFile(ctx, fileName, d.logger)
	// }

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
	if infoBlock.BlockID == 0 || *utils.StartBlockID > 0 {
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

		if *utils.EndBlockID > 0 && block.ID == *utils.EndBlockID {
			return nil
		}

		if *utils.StartBlockID == 0 || (*utils.StartBlockID > 0 && block.ID > *utils.StartBlockID) {
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
