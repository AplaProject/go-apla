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

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/notificator"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// BlockGenerator is daemon that generates blocks
func BlockGenerator(ctx context.Context, d *daemon) error {
	d.sleepTime = time.Second
	if service.IsNodePaused() {
		return nil
	}

	nodePosition, err := syspar.GetNodePositionByKeyID(conf.Config.KeyID)
	if err != nil {
		// we are not full node and can't generate new blocks
		d.sleepTime = 10 * time.Second
		d.logger.WithFields(log.Fields{"type": consts.JustWaiting, "error": err}).Debug("we are not full node, sleep for 10 seconds")
		return nil
	}

	DBLock()
	defer DBUnlock()

	// wee need fresh myNodePosition after locking
	nodePosition, err = syspar.GetNodePositionByKeyID(conf.Config.KeyID)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting node position by key id")
		return err
	}

	blockTimeCalculator, err := utils.BuildBlockTimeCalculator()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("building block time calculator")
		return err
	}

	timeToGenerate, err := blockTimeCalculator.SetClock(&utils.ClockWrapper{}).TimeToGenerate(nodePosition)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("calculating block time")
		return err
	}

	if !timeToGenerate {
		d.logger.WithFields(log.Fields{"type": consts.JustWaiting}).Debug("not my generation time")
		return nil
	}

	prevBlock := &model.InfoBlock{}
	_, err = prevBlock.Get()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting previous block")
		return err
	}

	NodePrivateKey, NodePublicKey, err := utils.GetNodeKeys()
	if err != nil || len(NodePrivateKey) < 1 {
		if err == nil {
			d.logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		}
		return err
	}

	dtx := DelayedTx{
		privateKey: NodePrivateKey,
		publicKey:  NodePublicKey,
		logger:     d.logger,
	}
	dtx.RunForBlockID(prevBlock.BlockID + 1)

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
				model.IncrementTxAttemptCount(nil, p.TxHash)
				break
			} else if err != nil {
				if err == parser.ErrLimitSkip {
					model.IncrementTxAttemptCount(nil, p.TxHash)
				} else {
					p.ProcessBadTransaction(err)
				}
				continue
			}
		}
		txList = append(txList, &trs[i])
	}

	// Block generation will be started only if we have transactions
	if len(trs) == 0 {
		return nil
	}

	header := &utils.BlockData{
		BlockID:      prevBlock.BlockID + 1,
		Time:         time.Now().Unix(),
		EcosystemID:  conf.Config.EcosystemID,
		KeyID:        conf.Config.KeyID,
		NodePosition: nodePosition,
		Version:      consts.BLOCK_VERSION,
	}

	blockBin, err := generateNextBlock(header, trs, NodePrivateKey, prevBlock.Hash)
	if err != nil {
		return err
	}

	err = parser.InsertBlockWOForks(blockBin, true)
	if err != nil {
		return err
	}

	go notificator.CheckTokenMovementLimits(nil, conf.Config.TokenMovement, header.BlockID)
	return nil
}

func generateNextBlock(blockHeader *utils.BlockData, trs []model.Transaction, key string, prevBlockHash []byte) ([]byte, error) {

	trData := make([][]byte, 0, len(trs))
	for _, tr := range trs {
		trData = append(trData, tr.Data)
	}

	return parser.MarshallBlock(blockHeader, trData, prevBlockHash, key)
}
