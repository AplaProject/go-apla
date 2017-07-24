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
 * Откат БД по блокам
 */
func (p *Parser) ParseDataRollback() error {
	var txType int
	p.dataPre()
	if p.dataType != 0 { // парсим только блоки
		// parse only blocks
		return utils.ErrInfo(fmt.Errorf("incorrect dataType"))
	}
	var err error

	err = p.ParseBlock()
	if err != nil {
		return utils.ErrInfo(err)
	}
	if len(p.BinaryData) > 0 {
		// вначале нужно получить размеры всех тр-ий, чтобы пройтись по ним в обратном порядке
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
			// удалим тр-ию
			// remove the transaction
			converter.BytesShift(&binForSize, txSize)
			if len(binForSize) == 0 {
				break
			}
		}
		sizesSlice = converter.SliceReverse(sizesSlice)
		for i := 0; i < len(sizesSlice); i++ {
			// обработка тр-ий может занять много времени, нужно отметиться
			// processing of the transaction may take a lot of time, we need to be marked
			p.UpdDaemonTime(p.GoroutineName)
			// отделим одну транзакцию
			transactionBinaryData := converter.BytesShiftReverse(&p.BinaryData, sizesSlice[i])
			p.TxBinaryData = transactionBinaryData
			txType = int(converter.BinToDecBytesShift(&p.TxBinaryData, 1))
			// узнаем кол-во байт, которое занимает размер и удалим размер
			// we'll get know the quantaty of bytes which the size takes
			converter.BytesShiftReverse(&p.BinaryData, len(converter.EncodeLength(sizesSlice[i])))
			hash, err := crypto.Hash(transactionBinaryData)
			if err != nil {
				log.Fatal(err)
			}
			hash = converter.BinToHex(hash)
			p.TxHash = string(hash)

			affect, err := model.MarkTransactionUnusedAndUnverified([]byte(p.TxHash))
			if err != nil {
				logging.WriteSelectiveLog(err)
				return p.ErrInfo(err)
			}
			logging.WriteSelectiveLog("affect: " + converter.Int64ToStr(affect))
			_, err = model.DeleteLogTransactionsByHash([]byte(p.TxHash))
			if err != nil {
				return p.ErrInfo(err)
			}
			// даем юзеру понять, что его тр-ия не в блоке
			// let user know that his territory isn't in the block
			ts := &model.TransactionsStatus{}
			err = ts.UpdateBlockID(0, []byte(p.TxHash))
			if err != nil {
				return p.ErrInfo(err)
			}
			// пишем тр-ию в очередь на проверку, авось пригодится
			// put the transaction in the turn for checking suddenly we will need it
			dataHex := converter.BinToHex(transactionBinaryData)
			log.Debug("DELETE FROM queue_tx WHERE hex(hash) = %s", p.TxHash)
			_, err = model.DeleteQueueTxByHash([]byte(p.TxHash))
			if err != nil {
				return p.ErrInfo(err)
			}
			queueTx := &model.QueueTx{Hash: []byte(p.TxHash), Data: []byte(dataHex)}
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
