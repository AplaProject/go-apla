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
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

func BlockGenerator(d *daemon, ctx context.Context) error {
	d.sleepTime = time.Second

	config := &model.Config{}
	if _, err := config.Get(); err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("cannot get config")
		return err
	}

	fullNodes := &model.FullNode{}
	_, err := fullNodes.FindNode(config.StateID, config.DltWalletID, config.StateID, config.DltWalletID)
	if err != nil || fullNodes.ID == 0 {
		// we are not full node and can't generate new blocks
		d.sleepTime = 10 * time.Second
		d.logger.WithFields(log.Fields{"type": consts.JustWaiting, "error": err}).Warning("we are not full node, sleep for 10 seconds")
		return nil
	}

	DBLock()
	defer DBUnlock()

	prevBlock := &model.InfoBlock{}
	_, err = prevBlock.Get()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting previous block")
		return err
	}

	// calculate the next block generation time
	sleepTime, err := model.GetSleepTime(config.DltWalletID, config.StateID, config.StateID, config.DltWalletID)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting sleep time")
		return err
	}
	toSleep := int64(sleepTime) - (time.Now().Unix() - int64(prevBlock.Time))
	if toSleep > 0 {
		d.logger.WithFields(log.Fields{"type": consts.JustWaiting, "seconds": toSleep}).Debug("sleeping n seconds")
		d.sleepTime = time.Duration(toSleep) * time.Second
		return nil
	}

	nodeKey := &model.MyNodeKey{}
	err = nodeKey.GetNodeWithMaxBlockID()
	if err != nil || len(nodeKey.PrivateKey) < 1 {
		if err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting node with max blockID")
		}
		d.logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		return err
	}

	p := new(parser.Parser)

	// verify transactions
	err = p.AllTxParser()
	if err != nil {
		return err
	}

	trs, err := model.GetAllUnusedTransactions()
	if err != nil || trs == nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all unused transactions")
		return err
	}

	blockBin, err := generateNextBlock(prevBlock, *trs, nodeKey.PrivateKey, config, time.Now().Unix())
	if err != nil {
		return err
	}

	err = parser.InsertBlockWOForks(blockBin)
	if err != nil {
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
