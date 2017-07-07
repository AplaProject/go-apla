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

	"context"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func BlockGenerator(d *daemon, ctx context.Context) error {
	d.sleepTime = time.Second

	locked, err := d.DbLock(ctx, d.goRoutineName)
	if !locked || err != nil {
		return err
	}
	defer d.DbUnlock(d.goRoutineName)

	config := &model.Config{}
	if err = config.GetConfig(); err != nil {
		return err
	}

	if config.StateID > 0 {
		systemState := &model.SystemRecognizedStates{}
		delegated, err := systemState.IsDelegated(config.StateID)
		if err != nil {
			return err
		}

		if delegated {
			// we are the state and we have delegated the node maintenance to another user or state
			d.sleepTime = 3600 * time.Second
			return nil
		}
	}

	fullNodes := &model.FullNodes{}
	if err = fullNodes.FindNode(config.StateID, config.DltWalletID, config.StateID, config.DltWalletID); err != nil {
		return err
	}

	if fullNodes.ID == 0 {
		// we are not full node and can't generate new blocks
		d.sleepTime = 10 * time.Second
		return nil
	}

	prevBlock := &model.InfoBlock{}
	err = prevBlock.GetInfoBlock()
	if err != nil {
		return err
	}

	// calculate the next block generation time
	sleepTime, err := d.GetSleepTime(config.DltWalletID, config.StateID, config.StateID, config.DltWalletID)
	if err != nil {
		return err
	}
	toSleep := int64(sleepTime) - (time.Now().Unix() - int64(prevBlock.Time))
	if toSleep > 0 {
		d.sleepTime = time.Duration(toSleep) * time.Second
		return nil
	}

	nodeKey := &model.MyNodeKeys{}
	err = nodeKey.GetNodeWithMaxBlockID()
	if err != nil || len(nodeKey.PrivateKey) < 1 {
		return err
	}

	p := new(parser.Parser)
	p.DCDB = d.DCDB

	// verify transactions
	err = p.AllTxParser()
	if err != nil {
		return err
	}

	trs, err := model.GetAllUnusedTransactions()
	if err != nil || trs == nil {
		return err
	}

	blockBin, err := generateNextBlock(prevBlock, *trs, string(nodeKey.PrivateKey), config, time.Now().Unix())
	if err != nil {
		return err
	}

	p.BinaryData = blockBin
	err = p.ParseDataFull(true)
	if err != nil {
		p.BlockError(err)
		return err
	}

	return nil
}

func generateNextBlock(prevBlock *model.InfoBlock, trs []model.Transactions, key string, c *model.Config, blockTime int64) ([]byte, error) {
	newBlockID := prevBlock.BlockID + 1

	var mrklArray [][]byte
	var blockDataTx []byte
	for _, tr := range trs {
		doubleHash, err := crypto.DoubleHash(tr.Data)
		if err != nil {
			return nil, err
		}
		mrklArray = append(mrklArray, converter.BinToHex(doubleHash)) // TODO: check it !!!!
		blockDataTx = append(blockDataTx, converter.EncodeLengthPlusData([]byte(tr.Data))...)
	}

	if len(mrklArray) == 0 {
		mrklArray = append(mrklArray, []byte("0"))
	}
	mrklRoot := utils.MerkleTreeRoot(mrklArray)

	forSign := fmt.Sprintf("0,%v,%v,%v,%v,%v,%s",
		newBlockID, prevBlock.Hash, blockTime, c.DltWalletID, c.StateID, string(mrklRoot))

	signed, err := crypto.Sign(key, forSign)
	if err != nil {
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
