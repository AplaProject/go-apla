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
	"bytes"
	"fmt"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"context"

	"encoding/hex"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func BlockGenerator(d *daemon, ctx context.Context) error {
	logger.LogDebug(consts.FuncStarted, "")
	d.sleepTime = time.Second

	locked, err := DbLock(ctx, d.goRoutineName)
	if !locked || err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}
	defer DbUnlock(d.goRoutineName)

	config := &model.Config{}
	if err = config.GetConfig(); err != nil {
		logger.LogError(consts.ConfigError, err)
		return err
	}

	if config.StateID > 0 {
		systemState := &model.SystemRecognizedState{}
		delegated, err := systemState.IsDelegated(config.StateID)
		if err == nil && delegated {
			// we are the state and we have delegated the node maintenance to another user or state
			logger.LogWarn(consts.JustWaiting, "we delegated block generation, sleep for hour")
			d.sleepTime = 3600 * time.Second
			return nil
		}
	}

	fullNodes := &model.FullNode{}
	err = fullNodes.FindNode(config.StateID, config.DltWalletID, config.StateID, config.DltWalletID)
	if err != nil || fullNodes.ID == 0 {
		// we are not full node and can't generate new blocks
		d.sleepTime = 10 * time.Second
		logger.LogWarn(consts.JustWaiting, "we are not full node, sleep for 10 seconds")
		return nil
	}

	prevBlock := &model.InfoBlock{}
	err = prevBlock.GetInfoBlock()
	if err != nil {
		logger.LogError(consts.BlockError, err)
		return err
	}

	// calculate the next block generation time
	sleepTime, err := model.GetSleepTime(config.DltWalletID, config.StateID, config.StateID, config.DltWalletID)
	if err != nil {
		logger.LogError(consts.DBError, fmt.Sprintf("can't get sleep time: %s", err))
		return err
	}
	toSleep := int64(sleepTime) - (time.Now().Unix() - int64(prevBlock.Time))
	if toSleep > 0 {
		logger.LogInfo(consts.JustWaiting, toSleep)
		d.sleepTime = time.Duration(toSleep) * time.Second
		return nil
	}

	nodeKey := &model.MyNodeKey{}
	err = nodeKey.GetNodeWithMaxBlockID()
	if err != nil || len(nodeKey.PrivateKey) < 1 {
		logger.LogError(consts.PrivateKeyError, err)
		return err
	}

	p := new(parser.Parser)

	// verify transactions
	err = p.AllTxParser()
	if err != nil {
		logger.LogError(consts.ParserError, err)
		return err
	}

	trs, err := model.GetAllUnusedTransactions()
	if err != nil || trs == nil {
		return err
	}
	logger.LogDebug(consts.DebugMessage, fmt.Sprintf("transactions to put in new block: %+v", trs))

	blockBin, err := generateNextBlock(prevBlock, *trs, hex.EncodeToString(nodeKey.PrivateKey), config, time.Now().Unix())
	if err != nil {
		logger.LogError(consts.BlockError, err)
		return err
	}

	p.BinaryData = blockBin
	logger.LogDebug(consts.DebugMessage, "try to parse new transactions")
	err = p.ParseDataFull(true)
	if err != nil {
		logger.LogError(consts.BlockError, err)
		p.BlockError(err)
		return err
	}

	return nil
}

func generateNextBlock(prevBlock *model.InfoBlock, trs []model.Transaction, key string, c *model.Config, blockTime int64) ([]byte, error) {
	logger.LogDebug(consts.FuncStarted, "")
	newBlockID := prevBlock.BlockID + 1

	var mrklArray [][]byte
	var blockDataTx []byte
	for _, tr := range trs {
		doubleHash, err := crypto.DoubleHash(tr.Data)
		if err != nil {
			logger.LogError(consts.CryptoError, err)
			return nil, err
		}
		mrklArray = append(mrklArray, converter.BinToHex(doubleHash))
		blockDataTx = append(blockDataTx, converter.EncodeLengthPlusData([]byte(tr.Data))...)
	}

	if len(mrklArray) == 0 {
		mrklArray = append(mrklArray, []byte("0"))
	}
	mrklRoot := utils.MerkleTreeRoot(mrklArray)

	forSign := fmt.Sprintf("0,%d,%s,%d,%d,%d,%s",
		newBlockID, prevBlock.Hash, blockTime, c.DltWalletID, c.StateID, mrklRoot)

	signed, err := crypto.Sign(key, forSign)
	if err != nil {
		logger.LogError(consts.CryptoError, err)
		return nil, err
	}

	var buf bytes.Buffer
	// fill header
	buf.Write(converter.DecToBin(0, 1))
	buf.Write(converter.DecToBin(newBlockID, 4))
	buf.Write(converter.DecToBin(blockTime, 4))
	buf.Write(converter.EncodeLenInt64InPlace(c.DltWalletID))
	buf.Write(converter.DecToBin(c.StateID, 1))
	buf.Write(converter.EncodeLengthPlusData(signed))
	// data
	buf.Write(blockDataTx)

	return buf.Bytes(), nil
}
