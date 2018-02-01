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
	"context"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

// BlockGenerator is daemon that generates blocks
func BlockGenerator(ctx context.Context, d *daemon) error {
	d.sleepTime = time.Second

	_, err := syspar.GetNodePositionByKeyID(conf.Config.KeyID)
	if err != nil {
		// we are not full node and can't generate new blocks
		d.sleepTime = 10 * time.Second
		d.logger.WithFields(log.Fields{"type": consts.JustWaiting, "error": err}).Debug("we are not full node, sleep for 10 seconds")
		return nil
	}

	DBLock()
	defer DBUnlock()

	// wee need fresh myNodePosition after locking
	myNodePosition, err := syspar.GetNodePositionByKeyID(conf.Config.KeyID)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting node position by key id")
		return err
	}

	prevBlock := &model.InfoBlock{}
	_, err = prevBlock.Get()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting previous block")
		return err
	}

	// calculate the next block generation time
	sleepTime, err := syspar.GetSleepTimeByKey(conf.Config.KeyID, converter.StrToInt64(prevBlock.NodePosition))
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

	NodePrivateKey, _, err := utils.GetNodeKeys()
	if err != nil || len(NodePrivateKey) < 1 {
		if err == nil {
			d.logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		}
		return err
	}

	p := new(parser.Parser)

	// verify transactions
	err = p.AllTxParser()
	if err != nil {
		return err
	}

	trs, err := model.GetAllUnusedTransactions()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all unused transactions")
		return err
	}
	if len(trs) == 0 {
		return nil
	}

	limits := parser.NewLimits(nil)
	// Checks preprocessing count limits
	txList := make([]*model.Transaction, 0, len(trs))
	for i, txItem := range trs {
		bufTransaction := bytes.NewBuffer(txItem.Data)
		p, err := parser.ParseTransaction(bufTransaction)
		if err != nil {
			p.ProcessBadTransaction(err)
			continue
		}
		if p.TxSmart != nil {
			err = limits.CheckLimit(p)
			if err == parser.ErrLimitStop && i > 0 {
				model.IncrementTxAttemptCount(p.TxHash)
				break
			} else if err != nil {
				if err == parser.ErrLimitSkip {
					model.IncrementTxAttemptCount(p.TxHash)
				} else {
					p.ProcessBadTransaction(err)
				}
				continue
			}
		}
		txList = append(txList, &trs[i])
	}

	blockBin, err := generateNextBlock(
		prevBlock,
		txList,
		NodePrivateKey,
		time.Now().Unix(),
		myNodePosition,
		conf.Config.EcosystemID,
		conf.Config.KeyID,
	)
	if err != nil {
		return err
	}
	return parser.InsertBlockWOForks(blockBin, true)
}

func generateNextBlock(
	prevBlock *model.InfoBlock,
	trs []*model.Transaction,
	key string,
	blockTime int64,
	myNodePosition int64,
	ecosystemID int64,
	keyID int64,
) ([]byte, error) {

	header := &utils.BlockData{
		BlockID:      prevBlock.BlockID + 1,
		Time:         time.Now().Unix(),
		EcosystemID:  ecosystemID,
		KeyID:        keyID,
		NodePosition: myNodePosition,
		Version:      consts.BLOCK_VERSION,
	}

	trData := make([][]byte, 0, len(trs))
	for _, tr := range trs {
		trData = append(trData, tr.Data)
	}

	return parser.MarshallBlock(header, trData, prevBlock.Hash, key)
}
