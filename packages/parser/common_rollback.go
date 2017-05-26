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
	//	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

//  если в ходе проверки тр-ий возникает ошибка, то вызываем откатчик всех занесенных тр-ий
// if the error appears during the checking of transactions, call the rollback of transactions
func (p *Parser) RollbackTo(binaryData []byte, skipCurrent bool) error {
	var err error
	if len(binaryData) > 0 {
		// вначале нужно получить размеры всех тр-ий, чтобы пройтись по ним в обратном порядке
		// in the beggining it's neccessary to obtain the sizes of all transactions in order to go through them in reverse order
		binForSize := binaryData
		var sizesSlice []int64
		for {
			txSize := utils.DecodeLength(&binForSize)
			if txSize == 0 {
				break
			}
			sizesSlice = append(sizesSlice, txSize)
			// удалим тр-ию
			// remove the transaction
			log.Debug("txSize", txSize)
			//log.Debug("binForSize", binForSize)
			utils.BytesShift(&binForSize, txSize)
			if len(binForSize) == 0 {
				break
			}
		}
		sizesSlice = utils.SliceReverse(sizesSlice)
		for i := 0; i < len(sizesSlice); i++ {
			// обработка тр-ий может занять много времени, нужно отметиться
			// processing of transaction may take a lot off time, we have to be marked
			p.UpdDaemonTime(p.GoroutineName)
			// отделим одну транзакцию
			// separate one transaction
			transactionBinaryData := utils.BytesShiftReverse(&binaryData, sizesSlice[i])
			transactionBinaryData_ := transactionBinaryData
			// узнаем кол-во байт, которое занимает размер и удалим размер
			// get to know the quantaty of bytes, which the size takes and remove it
			utils.BytesShiftReverse(&binaryData, len(lib.EncodeLength(sizesSlice[i])))
			p.TxHash = string(utils.Md5(transactionBinaryData))
			p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
			if err != nil {
				return utils.ErrInfo(err)
			}
			var (
				MethodName string
				err_       interface{}
			)
			if p.TxContract == nil {
				MethodName = consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
				p.TxMap = map[string][]byte{}
				err_ = utils.CallMethod(p, MethodName+"Init")
				if _, ok := err_.(error); ok {
					return utils.ErrInfo(err_.(error))
				}
			}
			// если дошли до тр-ии, которая вызвала ошибку, то откатываем только фронтальную проверку
			// if we get to the transaction, which caused the error, then we roll back only the frontal check
			/*if i == 0 {
						/*if skipCurrent { // тр-ия, которая вызвала ошибку закончилась еще до фронт. проверки, т.е. откатывать по ней вообще нечего
			// transaction that caused the error was finished before frontal check, then there is nothing to rall back
							continue
						}*/
			/*// если успели дойти только до половины фронтальной функции
			// If we reached only half of the frontal function
						MethodNameRollbackFront := MethodName + "RollbackFront"
						// откатываем только фронтальную проверку
			// roll back only frontal check
						err_ = utils.CallMethod(p, MethodNameRollbackFront)
						if _, ok := err_.(error); ok {
							return utils.ErrInfo(err_.(error))
						}*/
			/*} else if onlyFront {*/
			/*err_ = utils.CallMethod(p, MethodName+"RollbackFront")
			if _, ok := err_.(error); ok {
				return utils.ErrInfo(err_.(error))
			}*/
			/*} else {*/
			/*err_ = utils.CallMethod(p, MethodName+"RollbackFront")
			if _, ok := err_.(error); ok {
				return utils.ErrInfo(err_.(error))
			}*/
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
					err_ = utils.CallMethod(p, MethodName+"Rollback")
					if _, ok := err_.(error); ok {
						return utils.ErrInfo(err_.(error))
					}
				}
				err = p.DelLogTx(transactionBinaryData_)
				if err != nil {
					log.Error("error: %v", err)
				}
				affect, err := p.ExecSQLGetAffect("DELETE FROM transactions WHERE hex(hash) = ?", p.TxHash)
				if err != nil {
					utils.WriteSelectiveLog(err)
					return utils.ErrInfo(err)
				}
				utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
			}

			utils.WriteSelectiveLog("UPDATE transactions SET used = 0, verified = 0 WHERE hex(hash) = " + string(p.TxHash))
			affect, err := p.ExecSQLGetAffect("UPDATE transactions SET used = 0, verified = 0 WHERE hex(hash) = ?", p.TxHash)
			if err != nil {
				utils.WriteSelectiveLog(err)
				return utils.ErrInfo(err)
			}
			utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))

		}
	}
	return err
}
