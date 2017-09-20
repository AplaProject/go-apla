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

package parser

import (
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/logging"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

/**
 * Block rollback
 */
func (p *Parser) ParseDataRollback() error {
	var txType int
	p.dataPre()
	if p.dataType != 0 {
		// parse only blocks
		return utils.ErrInfo(fmt.Errorf("incorrect dataType"))
	}
	var err error

	err = p.ParseBlock()
	if err != nil {
		return utils.ErrInfo(err)
	}
	if len(p.BinaryData) > 0 {
		// in the beginning it is necessary to obtain the sizes of all the transactions in order to go through them in reverse order
		binForSize := p.BinaryData
		var sizesSlice []int64
		for {
			txSize, err := converter.DecodeLength(&binForSize)
			if err != nil {
				log.Fatal(err)
			}
			if txSize == 0 {
				break
			}
			sizesSlice = append(sizesSlice, txSize)
			// remove the transaction
			converter.BytesShift(&binForSize, txSize)
			if len(binForSize) == 0 {
				break
			}
		}
		sizesSlice = converter.SliceReverse(sizesSlice)
		for i := 0; i < len(sizesSlice); i++ {
			transactionBinaryData := converter.BytesShiftReverse(&p.BinaryData, sizesSlice[i])
			p.TxBinaryData = transactionBinaryData
			// get transaction type
			txType = int(converter.BinToDecBytesShift(&p.TxBinaryData, 1))
			// get transaction size
			converter.BytesShiftReverse(&p.BinaryData, len(converter.EncodeLength(sizesSlice[i])))
			hash, err := crypto.Hash(transactionBinaryData)
			if err != nil {
				return err
			}
			p.TxHash = hash

			affect, err := model.MarkTransactionUnusedAndUnverified(p.TxHash)
			if err != nil {
				logging.WriteSelectiveLog(err)
				return p.ErrInfo(err)
			}
			logging.WriteSelectiveLog("affect: " + converter.Int64ToStr(affect))
			_, err = model.DeleteLogTransactionsByHash(p.TxHash)
			if err != nil {
				return p.ErrInfo(err)
			}

			// let user know that his territory isn't in the block
			ts := &model.TransactionStatus{}
			err = ts.UpdateBlockID(0, p.TxHash)
			if err != nil {
				return p.ErrInfo(err)
			}
			// put the transaction in the turn for checking suddenly we will need it
			_, err = model.DeleteQueueTxByHash(p.TxHash)
			if err != nil {
				return p.ErrInfo(err)
			}
			queueTx := &model.QueueTx{Hash: p.TxHash, Data: transactionBinaryData}
			err = queueTx.Save()
			if err != nil {
				return p.ErrInfo(err)
			}

			p.TxSlice, _, err = p.ParseTransaction(&transactionBinaryData)
			if err != nil {
				return p.ErrInfo(err)
			}
			if p.TxContract != nil {
				if err := p.CallContract(smart.CallInit | smart.CallRollback); err != nil {
					return utils.ErrInfo(err)
				}
				if err = p.autoRollback(); err != nil {
					return p.ErrInfo(err)
				}
			} else {
				MethodName := consts.TxTypes[txType]
				parser, err := GetParser(p, MethodName)
				if err != nil {
					return p.ErrInfo(err)
				}
				result := parser.Init()
				if _, ok := result.(error); ok {
					return p.ErrInfo(result.(error))
				}
				result = parser.Rollback()
				if _, ok := result.(error); ok {
					return p.ErrInfo(result.(error))
				}
			}
		}
	}
	return nil
}
