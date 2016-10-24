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

	"github.com/EGaaS/go-mvp/packages/consts"
	"github.com/EGaaS/go-mvp/packages/smart"
	"github.com/EGaaS/go-mvp/packages/utils"
)

/**
фронт. проверка + занесение данных из блока в таблицы и info_block
*/
func (p *Parser) ParseDataFull(blockGenerator bool) error {

	p.dataPre()
	if p.dataType != 0 { // парсим только блоки
		return utils.ErrInfo(fmt.Errorf("incorrect dataType"))
	}
	var err error

	//if len(p.BinaryData) > 500000 {
	//	ioutil.WriteFile("block-"+string(utils.DSha256(p.BinaryData)), p.BinaryData, 0644)
	//}

	if blockGenerator {
		err = p.GetInfoBlock()
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	err = p.ParseBlock()
	if err != nil {
		return utils.ErrInfo(err)
	}

	// проверим данные, указанные в заголовке блока
	err = p.CheckBlockHeader()
	if err != nil {
		return utils.ErrInfo(err)
	}

	utils.WriteSelectiveLog("DELETE FROM transactions WHERE used = 1")
	afect, err := p.ExecSqlGetAffect("DELETE FROM transactions WHERE used = 1")
	if err != nil {
		utils.WriteSelectiveLog(err)
		return utils.ErrInfo(err)
	}
	utils.WriteSelectiveLog("afect: " + utils.Int64ToStr(afect))

	txCounter := make(map[int64]int64)
	p.fullTxBinaryData = p.BinaryData
	var txForRollbackTo []byte
	if len(p.BinaryData) > 0 {
		for {
			// обработка тр-ий может занять много времени, нужно отметиться
			p.UpdDaemonTime(p.GoroutineName)
			log.Debug("&p.BinaryData", p.BinaryData)
			transactionSize := utils.DecodeLength(&p.BinaryData)
			if len(p.BinaryData) == 0 {
				return utils.ErrInfo(fmt.Errorf("empty BinaryData"))
			}

			// отчекрыжим одну транзакцию от списка транзакций
			//log.Debug("++p.BinaryData=%x\n", p.BinaryData)
			//log.Debug("transactionSize", transactionSize)
			transactionBinaryData := utils.BytesShift(&p.BinaryData, transactionSize)
			transactionBinaryDataFull := transactionBinaryData
			//ioutil.WriteFile("/tmp/dctx", transactionBinaryDataFull, 0644)
			//ioutil.WriteFile("/tmp/dctxhash", utils.Md5(transactionBinaryDataFull), 0644)
			// добавляем взятую тр-ию в набор тр-ий для RollbackTo, в котором пойдем в обратном порядке
			txForRollbackTo = append(txForRollbackTo, utils.EncodeLengthPlusData(transactionBinaryData)...)
			//log.Debug("transactionBinaryData: %x\n", transactionBinaryData)
			//log.Debug("txForRollbackTo: %x\n", txForRollbackTo)

			err = p.CheckLogTx(transactionBinaryDataFull, false, false)
			if err != nil {
				err0 := p.RollbackTo(txForRollbackTo, true)
				if err0 != nil {
					log.Error("error: %v", err0)
				}
				return utils.ErrInfo(err)
			}

			utils.WriteSelectiveLog("UPDATE transactions SET used=1 WHERE hex(hash) = " + string(utils.Md5(transactionBinaryDataFull)))
			affect, err := p.ExecSqlGetAffect("UPDATE transactions SET used=1 WHERE hex(hash) = ?", utils.Md5(transactionBinaryDataFull))
			if err != nil {
				utils.WriteSelectiveLog(err)
				utils.WriteSelectiveLog("RollbackTo")
				err0 := p.RollbackTo(txForRollbackTo, true)
				if err0 != nil {
					log.Error("error: %v", err0)
				}
				return utils.ErrInfo(err)
			}
			utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
			//log.Debug("transactionBinaryData", transactionBinaryData)
			p.TxHash = string(utils.Md5(transactionBinaryData))
			log.Debug("p.TxHash %s", p.TxHash)
			p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
			log.Debug("p.TxSlice %v", p.TxSlice)
			if err != nil {
				err0 := p.RollbackTo(txForRollbackTo, true)
				if err0 != nil {
					log.Error("error: %v", err0)
				}
				return err
			}

			if p.BlockData.BlockId > 1 && p.TxContract == nil {
				var userId int64
				// txSlice[3] могут подсунуть пустой
				if len(p.TxSlice) > 3 {
					if !utils.CheckInputData(p.TxSlice[3], "int64") {
						return utils.ErrInfo(fmt.Errorf("empty user_id"))
					} else {
						userId = utils.BytesToInt64(p.TxSlice[3])
					}
				} else {
					return utils.ErrInfo(fmt.Errorf("empty user_id"))
				}

				// считаем по каждому юзеру, сколько в блоке от него транзакций
				txCounter[userId]++

				// чтобы 1 юзер не смог прислать дос-блок размером в 10гб, который заполнит своими же транзакциями
				if txCounter[userId] > consts.MAX_BLOCK_USER_TXS {
					err0 := p.RollbackTo(txForRollbackTo, true)
					if err0 != nil {
						log.Error("error: %v", err0)
					}
					return utils.ErrInfo(fmt.Errorf("max_block_user_transactions"))
				}
			}
			if p.TxContract == nil {
				// время в транзакции не может быть больше, чем на MAX_TX_FORW сек времени блока
				// и  время в транзакции не может быть меньше времени блока -24ч.
				if utils.BytesToInt64(p.TxSlice[2])-consts.MAX_TX_FORW > p.BlockData.Time || utils.BytesToInt64(p.TxSlice[2]) < p.BlockData.Time-consts.MAX_TX_BACK {
					err0 := p.RollbackTo(txForRollbackTo, true)
					if err0 != nil {
						log.Error("error: %v", err0)
					}
					return utils.ErrInfo(fmt.Errorf("incorrect transaction time"))
				}

				// проверим, есть ли такой тип тр-ий
				_, ok := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
				if !ok {
					return utils.ErrInfo(fmt.Errorf("nonexistent type"))
				}
			} else {
				if int64(p.TxPtr.(*consts.TXHeader).Time)-consts.MAX_TX_FORW > p.BlockData.Time || int64(p.TxPtr.(*consts.TXHeader).Time) < p.BlockData.Time-consts.MAX_TX_BACK {
					return utils.ErrInfo(fmt.Errorf("incorrect transaction time"))
				}

			}
			p.TxMap = map[string][]byte{}

			p.TxIds++

			if p.TxContract != nil {
				if err := p.CallContract(smart.CALL_INIT | smart.CALL_FRONT | smart.CALL_MAIN); err != nil {
					if p.TxContract.Called == smart.CALL_FRONT || p.TxContract.Called == smart.CALL_MAIN {
						err0 := p.RollbackTo(txForRollbackTo, true)
						if err0 != nil {
							log.Error("error: %v", err0)
						}
					}
					return utils.ErrInfo(err)
				}
			} else {
				MethodName := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
				log.Debug("MethodName", MethodName+"Init")
				err_ := utils.CallMethod(p, MethodName+"Init")
				if _, ok := err_.(error); ok {
					log.Error("error: %v", err)
					return utils.ErrInfo(err_.(error))
				}

				log.Debug("MethodName", MethodName+"Front")
				err_ = utils.CallMethod(p, MethodName+"Front")
				if _, ok := err_.(error); ok {
					log.Error("error: %v", err_)
					err0 := p.RollbackTo(txForRollbackTo, true)
					if err0 != nil {
						log.Error("error: %v", err0)
					}
					return utils.ErrInfo(err_.(error))
				}

				log.Debug("MethodName", MethodName)
				err_ = utils.CallMethod(p, MethodName)
				if _, ok := err_.(error); ok {
					log.Error("error: %v", err_)
					err0 := p.RollbackTo(txForRollbackTo, false)
					if err0 != nil {
						log.Error("error: %v", err0)
					}
					return utils.ErrInfo(err_.(error))
				}
			}
			// даем юзеру понять, что его тр-ия попала в блок
			p.ExecSql("UPDATE transactions_status SET block_id = ? WHERE hex(hash) = ?", p.BlockData.BlockId, utils.Md5(transactionBinaryDataFull))
			log.Debug("UPDATE transactions_status SET block_id = %d WHERE hex(hash) = %s", p.BlockData.BlockId, utils.Md5(transactionBinaryDataFull))

			// Тут было time(). А значит если бы в цепочке блоков были блоки в которых были бы одинаковые хэши тр-ий, то ParseDataFull вернул бы error
			err = p.InsertInLogTx(transactionBinaryDataFull, utils.BytesToInt64(p.TxMap["time"]))
			if err != nil {
				return utils.ErrInfo(err)
			}

			if len(p.BinaryData) == 0 {
				break
			}
		}
	}
	if blockGenerator {
		p.UpdBlockInfo()
		p.InsertIntoBlockchain()
	} else {
		p.UpdBlockInfo()

	}
	return nil
}
