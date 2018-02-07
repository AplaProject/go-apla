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
package parser

import (
	"database/sql"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

// RollbackToBlockID rollbacks blocks till blockID
func (p *Parser) RollbackToBlockID(blockID int64) error {
	logger := p.GetLogger()
	_, err := model.MarkVerifiedAndNotUsedTransactionsUnverified()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("marking verified and not used transactions unverified")
		return p.ErrInfo(err)
	}

	limit := 1000
	// roll back our blocks
	for {
		block := &model.Block{}
		blocks, err := block.GetBlocks(blockID, int32(limit))
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting blocks")
			return p.ErrInfo(err)
		}
		if len(blocks) == 0 {
			break
		}
		for _, block := range blocks {
			// roll back our blocks to the block blockID
			err = BlockRollback(block.Data)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
		blocks = blocks[:0]
	}
	block := &model.Block{}
	_, err = block.Get(blockID)
	if err != nil && err != sql.ErrNoRows {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block")
		return p.ErrInfo(err)
	}
	data := block.Data
	converter.BytesShift(&data, 1)
	iblock := converter.BinToDecBytesShift(&data, 4)
	time := converter.BinToDecBytesShift(&data, 4)
	ecosystemID := converter.BinToDecBytesShift(&data, 4)
	keyID := converter.BinToDecBytesShift(&data, 8)
	nodePosition := converter.BinToDecBytesShift(&data, 1)
	ib := &model.InfoBlock{
		Hash:         block.Hash,
		BlockID:      iblock,
		Time:         time,
		EcosystemID:  ecosystemID,
		KeyID:        keyID,
		NodePosition: converter.Int64ToStr(nodePosition)}

	err = ib.Update(p.DbTransaction)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating info block")
		return p.ErrInfo(err)
	}

	return nil
}
