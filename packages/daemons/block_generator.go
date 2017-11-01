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
	"time"

	"context"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/utils"
)

func BlockGenerator(d *daemon, ctx context.Context) error {
	d.sleepTime = time.Second

	config := &model.Config{}
	if _, err := config.Get(); err != nil {
		return err
	}

	fullNodes := &model.FullNode{}
	_, err := fullNodes.FindNode(config.StateID, config.DltWalletID, config.StateID, config.DltWalletID)
	if err != nil || fullNodes.ID == 0 {
		// we are not full node and can't generate new blocks
		d.sleepTime = 10 * time.Second
		log.Infof("we are not full node, sleep for 10 seconds")
		return nil
	}

	DBLock()
	defer DBUnlock()

	prevBlock := &model.InfoBlock{}
	_, err = prevBlock.Get()
	if err != nil {
		log.Errorf("can't get block: %s", err)
		return err
	}

	// calculate the next block generation time
	sleepTime, err := model.GetSleepTime(config.DltWalletID, config.StateID, config.StateID, config.DltWalletID)
	if err != nil {
		log.Errorf("can't get sleep time: %s", err)
		return err
	}
	toSleep := int64(sleepTime) - (time.Now().Unix() - int64(prevBlock.Time))
	if toSleep > 0 {
		log.Debugf("we need to sleep %d seconds to generate new block", toSleep)
		d.sleepTime = time.Duration(toSleep) * time.Second
		return nil
	}

	nodeKey := &model.MyNodeKey{}
	err = nodeKey.GetNodeWithMaxBlockID()
	if err != nil || len(nodeKey.PrivateKey) < 1 {
		log.Errorf("bad node private key: %s", err)
		return err
	}

	p := new(parser.Parser)

	// verify transactions
	err = p.AllTxParser()
	if err != nil {
		log.Errorf("transactions parser error: %s", err)
		return err
	}

	trs, err := model.GetAllUnusedTransactions()
	if err != nil || trs == nil {
		return err
	}
	log.Debugf("transactions to put in new block: %+v", trs)

	blockBin, err := generateNextBlock(prevBlock, *trs, nodeKey.PrivateKey, config, time.Now().Unix())
	if err != nil {
		log.Errorf("can't generate block: %s", err)
		return err
	}

	log.Debugf("try to parse new transactions")
	err = parser.InsertBlock(blockBin)
	if err != nil {
		log.Errorf("parser block error: %s", err)
		return err
	}

	return nil
}

func generateNextBlock(prevBlock *model.InfoBlock, trs []model.Transaction, key string, c *model.Config, blockTime int64) ([]byte, error) {
	header := &utils.BlockData{
		BlockID:  prevBlock.BlockID + 1,
		Time:     time.Now().Unix(),
		WalletID: c.DltWalletID,
		StateID:  c.StateID,
		Version:  consts.BLOCK_VERSION,
	}

	trData := make([][]byte, 0, len(trs))
	for _, tr := range trs {
		trData = append(trData, tr.Data)
	}

	return parser.MarshallBlock(header, trData, prevBlock.Hash, key)
}
