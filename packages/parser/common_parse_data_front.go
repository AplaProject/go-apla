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
	/*"fmt"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"*/
)

/**
 * Занесение данных из блока в БД
 * используется только в candidateBlock_is_ready
 */
// Putting data from the block into the database
// Is used only in candidateBlock_is_ready

/*
func (p *Parser) ParseDataFront() error {

	p.TxIds = []string{}
	p.dataPre()
	if p.dataType == 0 {
		// инфа о предыдущем блоке (т.е. последнем занесенном)
// information about previous block (the last  added)
		err := p.GetInfoBlock()
		if err != nil {
			return p.ErrInfo(err)
		}

		utils.WriteSelectiveLog("DELETE FROM transactions WHERE used=1")
		affect, err := p.ExecSqlGetAffect("DELETE FROM transactions WHERE used = 1")
		if err != nil {
			utils.WriteSelectiveLog(err)
			return p.ErrInfo(err)
		}
		utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))

		// разбор блока
// parse of block
		err = p.ParseBlock()
		if err != nil {
			return utils.ErrInfo(err)
		}

		//меркель рут нужен для updblockinfo()
// MrklRoot here is needed for updblockinfo()
		p.MrklRoot, err = utils.GetMrklroot(p.BinaryData, false)
		if err != nil {
			return utils.ErrInfo(err)
		}
		if len(p.BinaryData) > 0 {

			log.Debug("len(p.BinaryData)", len(p.BinaryData))

			for {
				transactionSize := utils.DecodeLength(&p.BinaryData)
				if len(p.BinaryData) == 0 {
					return utils.ErrInfo(fmt.Errorf("empty BinaryData"))
				}

				log.Debug("transactionSize", transactionSize)

				// отчекрыжим одну транзакцию от списка транзакций
// separate one transaction from the list of transactions
				transactionBinaryData := utils.BytesShift(&p.BinaryData, transactionSize)

				transactionBinaryDataFull := transactionBinaryData

				p.TxHash = string(utils.Md5(transactionBinaryData))
				log.Debug("p.TxHash", p.TxHash)
				p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
				log.Debug("p.TxSlice", p.TxSlice)
				if err != nil {
					return utils.ErrInfo(err)
				}
				if p.TxContract == nil {
					// txSlice[4] могут подсунуть пустой
// txSlice[4] could slip the empty one
					if len(p.TxSlice) > 4 {
						if !utils.CheckInputData(p.TxSlice[3], "int64") || !utils.CheckInputData(p.TxSlice[4], "int64") {
							return utils.ErrInfo(fmt.Errorf("empty wallet_id or citizen_id"))
						}
					} else {
						return utils.ErrInfo(fmt.Errorf("empty user_id"))
					}

					// проверим, есть ли такой тип тр-ий
// check if such a type of transaction exists
					_, ok := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
					if !ok {
						return utils.ErrInfo(fmt.Errorf("nonexistent type"))
					}
				}
				p.TxMap = map[string][]byte{}

				// для статы
// for statistics
				p.TxIds = append(p.TxIds, string(p.TxSlice[1]))

				MethodName := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]

				if p.TxContract != nil {
					if err := p.TxContract.Call(smart.CallInit | smart.CallAction); err != nil {
						return utils.ErrInfo(err)
					}
				} else {
					log.Debug("MethodName", MethodName+"Init")
					err_ := utils.CallMethod(p, MethodName+"Init")
					if _, ok := err_.(error); ok {
						log.Debug("error: %v", err)
						return utils.ErrInfo(err_.(error))
					}
					log.Debug("MethodName", MethodName)
					err_ = utils.CallMethod(p, MethodName)
					if _, ok := err_.(error); ok {
						log.Debug("error: %v", err)
						return utils.ErrInfo(err_.(error))
					}
				}

				utils.WriteSelectiveLog("UPDATE transactions SET used=1 WHERE hex(hash) = " + string(utils.Md5(transactionBinaryDataFull)))
				affect, err := p.ExecSqlGetAffect("UPDATE transactions SET used=1 WHERE hex(hash) = ?", utils.Md5(transactionBinaryDataFull))
				if err != nil {
					utils.WriteSelectiveLog(err)
					return utils.ErrInfo(err)
				}
				utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))

				// даем юзеру понять, что его тр-ия попала в блок
// let user to know that his transaction got in the block
				err = p.ExecSql("UPDATE transactions_status SET block_id = ? WHERE hex(hash) = ?", p.BlockData.BlockId, utils.Md5(transactionBinaryDataFull))
				if err != nil {
					return utils.ErrInfo(err)
				}

				if len(p.BinaryData) == 0 {
					break
				}
			}
		}

		p.UpdBlockInfo()
		p.InsertIntoBlockchain()
	} else {
		return utils.ErrInfo(fmt.Errorf("incorrect type"))
	}

	return nil
}
*/
