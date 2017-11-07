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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/net/context/ctxhttp"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/static"
	"github.com/AplaProject/go-apla/packages/utils"
)

// BlocksCollection collects and parses blocks
func BlocksCollection(d *daemon, ctx context.Context) error {
	if err := initialLoad(d, ctx); err != nil {
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return blocksCollection(d, ctx)
}

func initialLoad(d *daemon, ctx context.Context) error {

	// check for initial load
	toLoad, err := needLoad()
	if err != nil {
		return err
	}

	if toLoad {
		log.Debugf("star first block loading")
		if err := model.UpdateConfig("current_load_clockchain", "file"); err != nil {
			return err
		}

		if err := firstLoad(ctx, d); err != nil {
			return err
		}
	}

	if err := model.UpdateConfig("current_load_clockchain", "nodes"); err != nil {
		return err
	}

	return nil
}

func blocksCollection(d *daemon, ctx context.Context) error {

	// TODO: ????? remove from all tables in some test mode ?????

	hosts := syspar.GetHosts()
	log.Debug("hosts: %v", hosts)

	// get a host with the biggest block id
	host, maxBlockID, err := chooseBestHost(ctx, hosts)
	if err != nil {
		return err
	}

	// update our chain till maxBlockID from the host
	if err := updateChain(ctx, d, host, maxBlockID); err != nil {
		return err
	}

	return nil
}

// best host is a host with the biggest last block ID
func chooseBestHost(ctx context.Context, hosts []string) (string, int64, error) {
	type blockAndHost struct {
		host    string
		blockID int64
		err     error
	}
	c := make(chan blockAndHost, len(hosts))

	var wg sync.WaitGroup
	for _, h := range hosts {
		if ctx.Err() != nil {
			return "", 0, ctx.Err()
		}
		wg.Add(1)

		go func(host string) {
			blockID, err := getHostBlockID(host)
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

func getHostBlockID(host string) (int64, error) {
	conn, err := utils.TCPConn(host)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	// get max block request
	_, err = conn.Write(converter.DecToBin(consts.DATA_TYPE_MAX_BLOCK_ID, 2))
	if err != nil {
		return 0, err
	}

	// response
	blockIDBin := make([]byte, 4)
	_, err = conn.Read(blockIDBin)
	if err != nil {
		return 0, err
	}

	return converter.BinToDec(blockIDBin), nil
}

// load from host all blocks from our last block to maxBlockID
func updateChain(ctx context.Context, d *daemon, host string, maxBlockID int64) error {

	DBLock()
	defer DBUnlock()

	log.Debug("host", host)
	log.Debug("maxBlockID", maxBlockID)
	// get current block id from our blockchain
	curBlock := &model.InfoBlock{}
	if _, err := curBlock.Get(); err != nil {
		return err
	}

	log.Debug("curBlock.BlockID", curBlock.BlockID)
	for blockID := curBlock.BlockID + 1; blockID <= maxBlockID; blockID++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		blockBin, err := utils.GetBlockBody(host, blockID, consts.DATA_TYPE_BLOCK_BODY)
		if err != nil {
			return err
		}

		block, err := parser.ProcessBlockWherePrevFromBlockchainTable(blockBin)
		if err != nil {
			// we got bad block and should ban this host
			banNode(host, err)
			return err
		}

		// hash compare could be failed in the case of fork
		hashMatched, err := block.CheckHash()
		if err != nil {
			log.Debug("%v", err)
			//banNode(host, err)
			//return err
		}

		if !hashMatched {
			log.Debug("!hashMatched")
			// it should be fork, replace our previous blocks to ones from the host
			err := parser.GetBlocks(blockID-1, host, "rollback_blocks_2", consts.DATA_TYPE_BLOCK_BODY)
			if err != nil {
				banNode(host, err)
				return err
			}
			log.Debug("GetBlocks OK")
		} else {
			log.Debug("hashMatched")
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

func downloadChain(ctx context.Context, fileName, url string) error {

	for i := 0; i < consts.DOWNLOAD_CHAIN_TRY_COUNT; i++ {
		loadCtx, cancel := context.WithTimeout(ctx, time.Duration(syspar.GetUpdFullNodesPeriod())*time.Second)
		defer cancel()

		blockchainSize, err := downloadToFile(loadCtx, url, fileName)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			continue
		}
		if blockchainSize > consts.BLOCKCHAIN_SIZE {
			return nil
		}
	}
	return fmt.Errorf("can't download blockchain from %s", url)
}

// init first block from file or from embedded value
func loadFirstBlock() error {
	var newBlock []byte
	var err error

	if len(*utils.FirstBlockDir) > 0 {
		fileName := *utils.FirstBlockDir + "/1block"
		log.Debugf("load first block from file: %s", fileName)
		newBlock, _ = ioutil.ReadFile(fileName)
	} else {
		log.Debugf("load from assets")
		newBlock, err = static.Asset("static/1block")
		if err != nil {
			return err
		}
	}

	log.Infof("start to insert first block")
	if err = parser.InsertBlockWOForks(newBlock); err != nil {
		log.Errorf("failed to parse first block: %s", err)
		return err
	}

	log.Infof("first block inserted")
	return nil
}

func firstLoad(ctx context.Context, d *daemon) error {

	DBLock()
	defer DBUnlock()

	nodeConfig := &model.Config{}
	_, err := nodeConfig.Get()
	if err != nil {
		return err
	}

	if nodeConfig.FirstLoadBlockchain == "file" {
		log.Debugf("first load from file")
		blockchainURL := nodeConfig.FirstLoadBlockchainURL
		if len(blockchainURL) == 0 {
			blockchainURL = syspar.GetBlockchainURL()
		}

		fileName := *utils.Dir + "/public/blockchain"
		err = downloadChain(ctx, fileName, blockchainURL)
		if err != nil {
			return err
		}

		err = loadFromFile(ctx, fileName)
		if err != nil {
			return err
		}
	} else {
		err = loadFirstBlock()
	}

	return err
}

func needLoad() (bool, error) {
	infoBlock := &model.InfoBlock{}
	_, err := infoBlock.Get()
	if err != nil {
		return false, err
	}
	// we have empty blockchain, we need to load blockchain from file or other source
	if infoBlock.BlockID == 0 || *utils.StartBlockID > 0 {
		return true, nil
	}
	return false, nil
}

func banNode(host string, err error) {
	// TODO
}

func loadFromFile(ctx context.Context, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		block, err := readBlock(file)
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
func downloadToFile(ctx context.Context, url, file string) (int64, error) {
	resp, err := ctxhttp.Get(ctx, &http.Client{}, url)
	if err != nil {
		return 0, utils.ErrInfo(err)
	}
	defer resp.Body.Close()

	f, err := os.Create(file)
	if err != nil {
		return 0, utils.ErrInfo(err)
	}
	defer f.Close()

	var offset int64
	for {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}

		data, err := ioutil.ReadAll(io.LimitReader(resp.Body, 10000))
		if err != nil {
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
