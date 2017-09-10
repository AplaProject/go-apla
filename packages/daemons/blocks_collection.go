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

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// BlocksCollection collects and parses blocks
func BlocksCollection(d *daemon, ctx context.Context) error {
	logger.LogDebug(consts.FuncStarted, "")
	if err := initialLoad(d, ctx); err != nil {
		logger.LogError(consts.BlockchainLoadError, err)
		return err
	}

	if ctx.Err() != nil {
		logger.LogError(consts.ContextError, ctx.Err())
		return ctx.Err()
	}

	return blocksCollection(d, ctx)
}

func initialLoad(d *daemon, ctx context.Context) error {
	logger.LogDebug(consts.FuncStarted, "")
	// check for initial load
	toLoad, err := needLoad()
	if err != nil {
		logger.LogError(consts.BlockchainLoadError, err)
		return err
	}

	if toLoad {
		logger.LogDebug(consts.DebugMessage, "star first block loading")
		if err := model.UpdateConfig("current_load_clockchain", "file"); err != nil {
			logger.LogError(consts.DBError, err)
			return err
		}

		parser := new(parser.Parser)
		parser.GoroutineName = d.goRoutineName

		if err := firstLoad(ctx, d, parser); err != nil {
			logger.LogError(consts.BlockchainLoadError, err)
			return err
		}
	}

	if err := model.UpdateConfig("current_load_clockchain", "nodes"); err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}

	return nil
}

func blocksCollection(d *daemon, ctx context.Context) error {
	logger.LogDebug(consts.FuncStarted, "")
	// TODO: ????? remove from all tables in some test mode ?????

	hosts, err := model.GetFullNodesHosts()
	if err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}

	// get a host with the biggest block id
	host, maxBlockID, err := chooseBestHost(ctx, hosts)
	if err != nil {
		logger.LogError(consts.ContextError, err)
		return err
	}

	// update our chain till maxBlockID from the host
	if err := updateChain(ctx, d, host, maxBlockID); err != nil {
		logger.LogError(consts.BlockchainLoadError, err)
		return err
	}

	return nil
}

// best host is a host with the biggest last block ID
func chooseBestHost(ctx context.Context, hosts []string) (string, int64, error) {
	logger.LogDebug(consts.FuncStarted, "")
	type blockAndHost struct {
		host    string
		blockID int64
		err     error
	}
	c := make(chan blockAndHost, len(hosts))

	var wg sync.WaitGroup
	for _, h := range hosts {
		if ctx.Err() != nil {
			logger.LogError(consts.ContextError, ctx.Err())
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
		}(GetHostPort(h))
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
	logger.LogDebug(consts.FuncStarted, "")
	conn, err := utils.TCPConn(host)
	if err != nil {
		logger.LogError(consts.ConnectionError, err)
		return 0, err
	}
	defer conn.Close()

	// get max block request
	_, err = conn.Write(converter.DecToBin(consts.DATA_TYPE_MAX_BLOCK_ID, 2))
	if err != nil {
		logger.LogError(consts.ConnectionError, err)
		return 0, err
	}

	// response
	blockIDBin := make([]byte, 4)
	_, err = conn.Read(blockIDBin)
	if err != nil {
		logger.LogError(consts.ConnectionError, err)
		return 0, err
	}

	return converter.BinToDec(blockIDBin), nil
}

// load from host all blocks from our last block to maxBlockID
func updateChain(ctx context.Context, d *daemon, host string, maxBlockID int64) error {
	logger.LogDebug(consts.FuncStarted, "")
	locked, err := DbLock(ctx, d.goRoutineName)
	if !locked || err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}
	defer DbUnlock(d.goRoutineName)

	// get current block id from our blockchain
	curBlock := &model.InfoBlock{}
	if err = curBlock.GetInfoBlock(); err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}

	parser := new(parser.Parser)
	parser.GoroutineName = d.goRoutineName

	for blockID := curBlock.BlockID + 1; blockID <= maxBlockID; blockID++ {
		UpdMainLock()

		if ctx.Err() != nil {
			logger.LogError(consts.ContextError, err)
			return ctx.Err()
		}

		blockBin, err := utils.GetBlockBody(host, blockID, consts.DATA_TYPE_BLOCK_BODY)
		if err != nil {
			logger.LogError(consts.BlockError, err)
			return err
		}

		// get block header and block body
		header, body, err := parseBlock(blockID, blockBin)
		if err != nil || header == nil {
			// we got bad block and should ban this host
			banNode(host, err)
			logger.LogError(consts.NodeBan, fmt.Sprintf("%s %s", host, err))
			return err
		}

		// get hash of the previous block from OUR chain to check current block from host
		prevHash, err := getBlockHash(blockID - 1)
		if err != nil {
			logger.LogError(consts.BlockError, err)
			return err
		}

		// hash compare could be failed in the case of fork
		hashMatched, err := checkHash(*header, body, []byte(prevHash))
		if err != nil {
			logger.LogError(consts.BlockError, err)
			logger.LogError(consts.NodeBan, fmt.Sprintf("%s %s", host, err))
			banNode(host, err)
			return err
		}

		if !hashMatched {
			// it should be fork, replace our previous blocks to ones from the host
			err := parser.GetBlocks(blockID-1, host, "rollback_blocks_2",
				d.goRoutineName, consts.DATA_TYPE_BLOCK_BODY)

			if err != nil {
				logger.LogError(consts.BlockError, err)
				logger.LogError(consts.NodeBan, fmt.Sprintf("%s %s", host, err))
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

		parser.BinaryData = blockBin
		if err = parser.ParseDataFull(false); err != nil {
			banNode(host, err)
			parser.BlockError(err)
			logger.LogError(consts.BlockError, err)
			logger.LogError(consts.NodeBan, fmt.Sprintf("%s %s", host, err))
			return err
		}

		if err = parser.InsertIntoBlockchain(); err != nil {
			logger.LogError(consts.DBError, err)
			return err
		}
	}
	return nil
}

func downloadChain(ctx context.Context, fileName, url string) error {
	logger.LogDebug(consts.FuncStarted, "")
	for i := 0; i < consts.DOWNLOAD_CHAIN_TRY_COUNT; i++ {
		loadCtx, cancel := context.WithTimeout(ctx, time.Duration(syspar.GetUpdFullNodesPeriod())*time.Second)
		defer cancel()

		blockchainSize, err := downloadToFile(loadCtx, url, fileName)
		if err != nil {
			logger.LogError(consts.BlockchainLoadError, err)
			continue
		}
		if blockchainSize > consts.BLOCKCHAIN_SIZE {
			return nil
		}
	}
	logger.LogError(consts.BlockchainLoadError, fmt.Sprintf("can't download blockchain from %s", url))
	return fmt.Errorf("can't download blockchain from %s", url)
}

// init first block from file or from embedded value
func loadFirstBlock(parser *parser.Parser) error {
	logger.LogDebug(consts.FuncStarted, "")
	var newBlock []byte
	var err error

	if len(*utils.FirstBlockDir) > 0 {
		fileName := *utils.FirstBlockDir + "/1block"
		logger.LogDebug(consts.DebugMessage, fmt.Sprintf("load first block from file: %s", fileName))
		newBlock, _ = ioutil.ReadFile(fileName)
	} else {
		logger.LogDebug(consts.DebugMessage, fmt.Sprintf("load first block from assets"))
		newBlock, err = static.Asset("static/1block")
		if err != nil {
			return err
		}
	}
	parser.BinaryData = newBlock
	parser.CurrentVersion = consts.VERSION

	logger.LogDebug(consts.DebugMessage, "start to insert first block")
	if err = parser.ParseDataFull(false); err != nil {
		parser.BlockError(err)
		logger.LogError(consts.BlockError, err)
		return err
	}

	if err = parser.InsertIntoBlockchain(); err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}
	logger.LogDebug(consts.DebugMessage, "first block inserted")

	return nil
}

// check block and get block header and block body
func parseBlock(blockID int64, binaryBlock []byte) (header *utils.BlockData, body []byte, err error) {
	logger.LogDebug(consts.FuncStarted, "")
	if len(binaryBlock) == 0 {
		err = fmt.Errorf("block is empty")
		return
	}

	converter.BytesShift(&binaryBlock, 1) // remove 1-st byte - type (block/transaction)
	header = utils.ParseBlockHeader(&binaryBlock)

	if int64(len(binaryBlock)) > syspar.GetMaxBlockSize() {
		err = fmt.Errorf(`len(binaryBlock) > variables.Int64["max_block_size"]  %v > %v`,
			len(binaryBlock), syspar.GetMaxBlockSize())
		return
	}

	if header.BlockID != blockID {
		err = fmt.Errorf(`header.BlockId != blockID  %v > %v`, header.BlockID, blockID)
		return
	}

	body = binaryBlock
	return
}

// check if new block is from our chain
func checkHash(header utils.BlockData, body []byte, prevHash []byte) (bool, error) {
	if header.BlockID == 1 {
		return true, nil
	}

	mrklRoot, err := utils.GetMrklroot(body, false, syspar.GetMaxTxSize(), syspar.GetMaxTxCount())
	if err != nil {
		return true, err
	}

	// public key of those who has generated this block
	var nodePublicKey []byte
	if header.WalletID != 0 {
		wallet := &model.DltWallet{}
		err = wallet.GetWallet(header.WalletID)
		if err != nil {
			logger.LogError(consts.DBError, err)
			return true, err
		}
		nodePublicKey = []byte(wallet.PublicKey)
	} else {
		systemState := &model.SystemRecognizedState{}
		err = systemState.GetState(header.StateID)
		if err != nil {
			logger.LogError(consts.DBError, err)
			return true, err
		}
		nodePublicKey = systemState.NodePublicKey
	}

	// SIGN from 128 bytes to 512 bytes. Signature from TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, USER_ID, LEVEL, MRKL_ROOT
	forSign := fmt.Sprintf("0,%v,%v,%v,%v,%v,%s", header.BlockID, prevHash,
		header.Time, header.WalletID, header.StateID, mrklRoot)

	_, err = utils.CheckSign([][]byte{nodePublicKey}, forSign, header.Sign, true)
	if err != nil {
		// check failed, it seems that this new block is from the different fork
		return false, nil
	}

	return true, nil
}

func firstLoad(ctx context.Context, d *daemon, parser *parser.Parser) error {
	logger.LogDebug(consts.FuncStarted, "")
	locked, err := DbLock(ctx, d.goRoutineName)
	if !locked || err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}
	defer DbUnlock(d.goRoutineName)

	nodeConfig := &model.Config{}
	err = nodeConfig.GetConfig()
	if err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}

	if nodeConfig.FirstLoadBlockchain == "file" {
		logger.LogDebug(consts.DebugMessage, "first load from file")
		blockchainURL := nodeConfig.FirstLoadBlockchainURL
		if len(blockchainURL) == 0 {
			blockchainURL = syspar.GetBlockchainURL()
		}

		fileName := *utils.Dir + "/public/blockchain"
		err = downloadChain(ctx, fileName, blockchainURL)
		if err != nil {
			logger.LogError(consts.BlockchainLoadError, err)
			return err
		}

		err = loadFromFile(ctx, parser, fileName)
		if err != nil {
			logger.LogError(consts.BlockchainLoadError, err)
			return err
		}
	} else {
		err = loadFirstBlock(parser)
	}

	return err
}

func needLoad() (bool, error) {
	logger.LogDebug(consts.FuncStarted, "")
	infoBlock := &model.InfoBlock{}
	err := infoBlock.GetInfoBlock()
	if err != nil {
		logger.LogError(consts.DBError, err)
		return false, err
	}
	// we have empty blockchain, we need to load blockchain from file or other source
	if infoBlock.BlockID == 0 || *utils.StartBlockID > 0 {
		logger.LogDebug(consts.DebugMessage, "blockchain should be loaded")
		return true, nil
	}
	logger.LogDebug(consts.DebugMessage, "blockchain shouldn't be loaded")
	return false, nil
}
func getBlockHash(blockID int64) (string, error) {
	logger.LogDebug(consts.FuncStarted, "")
	if blockID > 1 {
		block := &model.Block{}
		err := block.GetBlock(blockID)
		if err != nil {
			logger.LogError(consts.DBError, err)
			return "", err
		}
		return string(converter.BinToHex(block.Hash)), nil
	}

	return "0", nil
}

func banNode(host string, err error) {
	// TODO
}

func loadFromFile(ctx context.Context, parser *parser.Parser, fileName string) error {
	logger.LogDebug(consts.FuncStarted, "")
	file, err := os.Open(fileName)
	if err != nil {
		logger.LogError(consts.IOError, err)
		return err
	}
	defer file.Close()

	first := true
	for {
		if ctx.Err() != nil {
			logger.LogError(consts.ContextError, err)
			return ctx.Err()
		}

		if err = UpdMainLock(); err != nil {
			logger.LogError(consts.DBError, err)
			return err
		}

		block, err := readBlock(file)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			logger.LogError(consts.IOError, err)
			return err
		}

		if block == nil {
			return nil
		}

		if *utils.EndBlockID > 0 && block.ID == *utils.EndBlockID {
			return nil
		}

		if *utils.StartBlockID == 0 || (*utils.StartBlockID > 0 && block.ID > *utils.StartBlockID) {
			parser.BinaryData = block.Data
			if first {
				parser.CurrentVersion = consts.VERSION
				first = false
			}

			if err = parser.ParseDataFull(false); err != nil {
				parser.BlockError(err)
				logger.LogError(consts.BlockError, err)
				return err
			}

			if err = parser.InsertIntoBlockchain(); err != nil {
				logger.LogError(consts.DBError, err)
				return err
			}
		}
	}
}

// downloadToFile downloads and saves the specified file
func downloadToFile(ctx context.Context, url, file string) (int64, error) {
	logger.LogDebug(consts.FuncStarted, "")
	resp, err := ctxhttp.Get(ctx, &http.Client{}, url)
	if err != nil {
		logger.LogError(consts.ContextError, err)
		return 0, utils.ErrInfo(err)
	}
	defer resp.Body.Close()

	f, err := os.Create(file)
	if err != nil {
		logger.LogError(consts.IOError, err)
		return 0, utils.ErrInfo(err)
	}
	defer f.Close()

	var offset int64
	for {
		if ctx.Err() != nil {
			logger.LogError(consts.ContextError, ctx.Err())
			return 0, ctx.Err()
		}

		data, err := ioutil.ReadAll(io.LimitReader(resp.Body, 10000))
		if err != nil {
			logger.LogError(consts.IOError, err)
			return offset, utils.ErrInfo(err)
		}

		f.WriteAt(data, offset)
		offset += int64(len(data))
		if len(data) == 0 {
			break
		}
		logger.LogDebug(consts.DebugMessage, fmt.Sprintf("read %s", url))
	}
	return offset, nil
}
