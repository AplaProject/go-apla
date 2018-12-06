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

	"github.com/AplaProject/go-apla/packages/block"
	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/notificator"
	"github.com/AplaProject/go-apla/packages/protocols"
	"github.com/AplaProject/go-apla/packages/service"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/utils"

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

	btc := protocols.NewBlockTimeCounter()
	at := time.Now()

	if exists, err := btc.BlockForTimeExists(at, int(nodePosition)); exists || err != nil {
		return nil
	}

	timeToGenerate, err := btc.TimeToGenerate(at, int(nodePosition))
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.BlockError, "error": err, "position": nodePosition}).Debug("calculating block time")
		return err
	}

	if !timeToGenerate {
		d.logger.WithFields(log.Fields{"type": consts.JustWaiting}).Debug("not my generation time")
		return nil
	}

	_, endTime, err := btc.RangeByTime(time.Now())
	if err != nil {
		log.WithFields(log.Fields{"type": consts.TimeCalcError, "error": err}).Error("on getting end time of generation")
	}

	done := time.After(endTime.Sub(time.Now()))
	prevBlock, _, found, err := blockchain.GetLastBlock(nil)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting previous block")
		return err
	}
	if !found {
		d.logger.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Error("previous block not found")
		return nil
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
	dtx.RunForBlockID(prevBlock.Header.BlockID + 1)

	err = transaction.ProcessTransactionsQueue()
	if err != nil {
		return err
	}
	txs, err := blockchain.GetTxsToProcess(nil)
	if err != nil {
		return err
	}
	trs, err := processTransactions(txs, d.logger, done)
	if err != nil {
		return err
	}

	// Block generation will be started only if we have transactions
	if len(trs) == 0 {
		return nil
	}
	hashes := [][]byte{}
	for _, tx := range trs {
		hsh, err := tx.Hash()
		if err != nil {
			return err
		}
		hashes = append(hashes, hsh)
	}

	header := &blockchain.BlockHeader{
		BlockID:      prevBlock.Header.BlockID + 1,
		Time:         time.Now().Unix(),
		EcosystemID:  0,
		KeyID:        conf.Config.KeyID,
		NodePosition: nodePosition,
		Version:      consts.BLOCK_VERSION,
	}
	bBlock := &blockchain.Block{
		Header:   header,
		TxHashes: hashes,
	}

	err = block.InsertBlockWOForks(bBlock, txs, true, false)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"Block": header.String(), "type": consts.SyncProcess}).Debug("Generated block ID")
	go notificator.CheckTokenMovementLimits(nil, conf.Config.TokenMovement, header.BlockID)

	return nil
}

func processTransactions(trs []*blockchain.Transaction, logger *log.Entry, done <-chan time.Time) ([]*blockchain.Transaction, error) {
	// verify transactions
	limits := block.NewLimits(nil)

	type badTxStruct struct {
		hash  []byte
		msg   string
		keyID int64
	}

	processBadTx := func() chan badTxStruct {
		ch := make(chan badTxStruct)

		go func() {
			for badTxItem := range ch {
				block.BadTxForBan(badTxItem.keyID)
				blockchain.SetTransactionError(nil, badTxItem.hash, badTxItem.msg)
			}
		}()

		return ch
	}

	processIncAttemptCnt := func() chan *blockchain.Transaction {
		ch := make(chan *blockchain.Transaction)
		go func() {
			for tx := range ch {
				hash, _ := tx.Hash()
				blockchain.IncrementTxAttemptCount(nil, hash)
			}
		}()

		return ch
	}

	txBadChan := processBadTx()
	attemptCountChan := processIncAttemptCnt()

	defer func() {
		close(txBadChan)
		close(attemptCountChan)
	}()

	// Checks preprocessing count limits
	txList := make([]*blockchain.Transaction, 0, len(trs))
	var err error
	for i, txItem := range trs {
		select {
		case <-done:
			return txList, err
		default:
			p, err := transaction.FromBlockchainTransaction(txItem)
			if err != nil {
				if p != nil {
					txBadChan <- badTxStruct{hash: p.TxHash, msg: err.Error(), keyID: p.TxHeader.KeyID}
				}
				continue
			}

			if err := p.Check(time.Now().Unix()); err != nil {
				txBadChan <- badTxStruct{hash: p.TxHash, msg: err.Error(), keyID: p.TxHeader.KeyID}
				continue
			}

			if p.TxSmart != nil {
				err = limits.CheckLimit(p)
				if err == block.ErrLimitStop && i > 0 {
					attemptCountChan <- txItem
					break
				} else if err != nil {
					if err == block.ErrLimitSkip {
						attemptCountChan <- txItem
					} else {
						txBadChan <- badTxStruct{hash: p.TxHash, msg: err.Error(), keyID: p.TxHeader.KeyID}
					}
					continue
				}
			}
			txList = append(txList, trs[i])
		}
	}
	return txList, nil
}
