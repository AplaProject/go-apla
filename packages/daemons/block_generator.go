//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package daemons

import (
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

	// Block generation will be started only if we have transactions
	if len(trs) == 0 {
		return nil
	}

	blockBin, err := generateNextBlock(
		prevBlock,
		trs,
		NodePrivateKey,
		time.Now().Unix(),
		myNodePosition,
		conf.Config.EcosystemID,
		conf.Config.KeyID,
	)
	if err != nil {
		return err
	}
	return parser.InsertBlockWOForks(blockBin)
}

func generateNextBlock(
	prevBlock *model.InfoBlock,
	trs []model.Transaction,
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
