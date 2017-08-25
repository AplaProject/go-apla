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
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/logging"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// RollbackTo rollbacks proceeded transactions
// if the error appears during the checking of transactions, call the rollback of transactions
func (p *Parser) RollbackTo(binaryData []byte, skipCurrent bool) error {
	var err error
	if len(binaryData) > 0 {
		// in the beggining it's neccessary to obtain the sizes of all transactions in order to go through them in reverse order
		binForSize := binaryData
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
			log.Debug("txSize", txSize)
			converter.BytesShift(&binForSize, txSize)
			if len(binForSize) == 0 {
				break
			}
		}
		sizesSlice = converter.SliceReverse(sizesSlice)
		for i := 0; i < len(sizesSlice); i++ {
			// processing of transaction may take a lot off time, we have to be marked
			// separate one transaction
			transactionBinaryData := converter.BytesShiftReverse(&binaryData, sizesSlice[i])
			binaryData := transactionBinaryData
			// get know the quantity of bytes, which the size takes and remove it
			converter.BytesShiftReverse(&binaryData, len(converter.EncodeLength(sizesSlice[i])))
			hash, err := crypto.Hash(transactionBinaryData)
			if err != nil {
				log.Fatal(err)
			}
			p.TxHash = hash
			p.TxBinaryData = transactionBinaryData
			converter.BinToDecBytesShift(&p.TxBinaryData, 1)
			p.TxSlice, _, err = p.ParseTransaction(&transactionBinaryData)
			if err != nil {
				return utils.ErrInfo(err)
			}
			var (
				MethodName string
				err_       interface{}
				parser     ParserInterface
			)
			if p.TxContract == nil {
				MethodName = consts.TxTypes[converter.BytesToInt(p.TxSlice[1])]
				parser, err = GetParser(p, MethodName)
				if err != nil {
					return utils.ErrInfo(err)
				}
				if parser != nil {
					p.TxMap = map[string][]byte{}
					err_ = parser.Init()
					if _, ok := err_.(error); ok {
						return utils.ErrInfo(err_.(error))
					}
				}
			}
			if (i == 0 && !skipCurrent) || i > 0 {
				log.Debug(MethodName + "Rollback")
				if p.TxContract != nil {
					if err := p.CallContract(smart.CallInit | smart.CallRollback); err != nil {
						return utils.ErrInfo(err)
					}
					if err = p.autoRollback(); err != nil {
						return p.ErrInfo(err)
					}
				} else {
					err_ = parser.Rollback()
					if _, ok := err_.(error); ok {
						return utils.ErrInfo(err_.(error))
					}
				}
				txHash, err := crypto.Hash(binaryData)
				if err != nil {
					p.ErrInfo(binaryData)
				}
				logTx := &model.LogTransaction{Hash: txHash}
				err = logTx.Delete()
				if err != nil {
					log.Error("error: %v", err)
				}
				affect, err := model.DeleteTransactionByHash(p.TxHash)
				if err != nil {
					logging.WriteSelectiveLog(err)
					return utils.ErrInfo(err)
				}
				logging.WriteSelectiveLog("affect: " + converter.Int64ToStr(affect))
			}

			affect, err := model.MarkTransactionUnusedAndUnverified(p.TxHash)
			if err != nil {
				logging.WriteSelectiveLog(err)
				return utils.ErrInfo(err)
			}
			logging.WriteSelectiveLog("affect: " + converter.Int64ToStr(affect))
		}
	}
	return err
}
