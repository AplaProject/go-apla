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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

//  если в ходе проверки тр-ий возникает ошибка, то вызываем откатчик всех занесенных тр-ий
func (p *Parser) RollbackTo(binaryData []byte, skipCurrent bool) error {
	var err error
	if len(binaryData) > 0 {
		// вначале нужно получить размеры всех тр-ий, чтобы пройтись по ним в обратном порядке
		binForSize := binaryData
		var sizesSlice []int64
		for {
			txSize := utils.DecodeLength(&binForSize)
			if txSize == 0 {
				break
			}
			sizesSlice = append(sizesSlice, txSize)
			// удалим тр-ию
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
			p.UpdDaemonTime(p.GoroutineName)
			// отделим одну транзакцию
			transactionBinaryData := utils.BytesShiftReverse(&binaryData, sizesSlice[i])
			transactionBinaryData_ := transactionBinaryData
			// узнаем кол-во байт, которое занимает размер
			size_ := len(utils.EncodeLength(sizesSlice[i]))
			// удалим размер
			utils.BytesShiftReverse(&binaryData, size_)
			p.TxHash = string(utils.Md5(transactionBinaryData))
			p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
			if err != nil {
				return utils.ErrInfo(err)
			}
			MethodName := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
			p.TxMap = map[string][]byte{}
			err_ := utils.CallMethod(p, MethodName+"Init")
			if _, ok := err_.(error); ok {
				return utils.ErrInfo(err_.(error))
			}

			// если дошли до тр-ии, которая вызвала ошибку, то откатываем только фронтальную проверку
			/*if i == 0 {
				/*if skipCurrent { // тр-ия, которая вызвала ошибку закончилась еще до фронт. проверки, т.е. откатывать по ней вообще нечего
					continue
				}*/
				/*// если успели дойти только до половины фронтальной функции
				MethodNameRollbackFront := MethodName + "RollbackFront"
				// откатываем только фронтальную проверку
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
				log.Debug(MethodName+"Rollback")
				err_ = utils.CallMethod(p, MethodName+"Rollback")
				if _, ok := err_.(error); ok {
					return utils.ErrInfo(err_.(error))
				}
				err = p.DelLogTx(transactionBinaryData_)
				if err!=nil {
					log.Error("error: %v", err)
				}
				affect, err := p.ExecSqlGetAffect("DELETE FROM transactions WHERE hex(hash) = ?", p.TxHash)
				if err != nil {
					utils.WriteSelectiveLog(err)
					return utils.ErrInfo(err)
				}
				utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
			}

				utils.WriteSelectiveLog("UPDATE transactions SET used = 0, verified = 0 WHERE hex(hash) = " + string(p.TxHash))
				affect, err := p.ExecSqlGetAffect("UPDATE transactions SET used = 0, verified = 0 WHERE hex(hash) = ?", p.TxHash)
				if err != nil {
					utils.WriteSelectiveLog(err)
					return utils.ErrInfo(err)
				}
				utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))

		}
	}
	return err
}
