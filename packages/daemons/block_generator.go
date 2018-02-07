// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package daemons

import (
	"bytes"
	"context"
	"time"

	"github.com/GenesisCommunity/go-genesis/packages/conf"

	"github.com/GenesisCommunity/go-genesis/packages/config/syspar"
	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/model"
	"github.com/GenesisCommunity/go-genesis/packages/parser"
	"github.com/GenesisCommunity/go-genesis/packages/utils"

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

	QueueParserBlocks(ctx, d)

	DBLock()
	defer DBUnlock()

	// wee need fresh myNodePosition after locking
	nodePosition, err = syspar.GetNodePositionByKeyID(conf.Config.KeyID)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting node position by key id")
		return err
	}

	blockTimeCalculator, err := utils.BuildBlockTimeCalculator(nil)
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

	trs, err := processTransactions(d.logger)
	if err != nil {
		return err
	}

	// Block generation will be started only if we have transactions
	if len(trs) == 0 {
		return nil
	}

	header := &utils.BlockData{
		BlockID:      prevBlock.BlockID + 1,
		Time:         time.Now().Unix(),
		EcosystemID:  0,
		KeyID:        conf.Config.KeyID,
		NodePosition: nodePosition,
		Version:      consts.BLOCK_VERSION,
	}

	timeToGenerate, err = blockTimeCalculator.SetClock(&utils.ClockWrapper{}).TimeToGenerate(nodePosition)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("calculating block time")
		return err
	}

	if !timeToGenerate {
		d.logger.WithFields(log.Fields{"type": consts.JustWaiting}).Debug("not my generation time")
		return nil
	}

	blockBin, err := generateNextBlock(header, trs, NodePrivateKey, prevBlock.Hash)
	if err != nil {
		return err
	}

	err = block.InsertBlockWOForks(blockBin, true, false)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"Block": header.String(), "type": consts.SyncProcess}).Error("Generated block ID")

	go notificator.CheckTokenMovementLimits(nil, conf.Config.TokenMovement, header.BlockID)
	return nil
}

func generateNextBlock(blockHeader *utils.BlockData, trs []*model.Transaction, key string, prevBlockHash []byte) ([]byte, error) {
	trData := make([][]byte, 0, len(trs))
	for _, tr := range trs {
		trData = append(trData, tr.Data)
	}

	return block.MarshallBlock(blockHeader, trData, prevBlockHash, key)
}

func processTransactions(logger *log.Entry) ([]*model.Transaction, error) {
	p := new(transaction.Transaction)

	// verify transactions
	err := transaction.ProcessTransactionsQueue(p.DbTransaction)
	if err != nil {
		return nil, err
	}

	trs, err := model.GetAllUnusedTransactions(syspar.GetMaxTxCount())
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all unused transactions")
		return nil, err
	}

	limits := block.NewLimits(nil)
	// Checks preprocessing count limits
	txList := make([]*model.Transaction, 0, len(trs))
	for i, txItem := range trs {
		bufTransaction := bytes.NewBuffer(txItem.Data)
		p, err := transaction.UnmarshallTransaction(bufTransaction)
		if err != nil {
			if p != nil {
				transaction.MarkTransactionBad(p.DbTransaction, p.TxHash, err.Error())
			}
			continue
		}

		if err := p.Check(time.Now().Unix(), false); err != nil {
			transaction.MarkTransactionBad(p.DbTransaction, p.TxHash, err.Error())
			continue
		}

		if p.TxSmart != nil {
			err = limits.CheckLimit(p)
			if err == block.ErrLimitStop && i > 0 {
				model.IncrementTxAttemptCount(nil, p.TxHash)
				break
			} else if err != nil {
				if err == block.ErrLimitSkip {
					model.IncrementTxAttemptCount(nil, p.TxHash)
				} else {
					transaction.MarkTransactionBad(p.DbTransaction, p.TxHash, err.Error())
				}
				continue
			}
		}
		txList = append(txList, trs[i])
	}

	return txList, nil
}
