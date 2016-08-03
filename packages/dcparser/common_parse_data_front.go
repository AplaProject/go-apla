package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

/**
 * Занесение данных из блока в БД
 * используется только в candidateBlock_is_ready
 */
func (p *Parser) ParseDataFront() error {

	p.TxIds = []string{}
	p.dataPre()
	if p.dataType == 0 {
		// инфа о предыдущем блоке (т.е. последнем занесенном)
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
		err = p.ParseBlock()
		if err != nil {
			return utils.ErrInfo(err)
		}

		p.Variables, err = p.GetAllVariables()
		if err != nil {
			return utils.ErrInfo(err)
		}
		//меркель рут нужен для updblockinfo()
		p.MrklRoot, err = utils.GetMrklroot(p.BinaryData, p.Variables, false)
		if err != nil {
			return utils.ErrInfo(err)
		}
		if len(p.BinaryData) > 0 {
			for {
				transactionSize := utils.DecodeLength(&p.BinaryData)
				if len(p.BinaryData) == 0 {
					return utils.ErrInfo(fmt.Errorf("empty BinaryData"))
				}
				// отчекрыжим одну транзакцию от списка транзакций
				transactionBinaryData := utils.BytesShift(&p.BinaryData, transactionSize)
				transactionBinaryDataFull := transactionBinaryData

				p.TxHash = string(utils.Md5(transactionBinaryData))
				log.Debug("p.TxHash", p.TxHash)
				p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
				log.Debug("p.TxSlice", p.TxSlice)
				if err != nil {
					return utils.ErrInfo(err)
				}

				// txSlice[3] могут подсунуть пустой
				if len(p.TxSlice) > 3 {
					if !utils.CheckInputData(p.TxSlice[3], "int64") {
						return utils.ErrInfo(fmt.Errorf("empty user_id"))
					}
				} else {
					return utils.ErrInfo(fmt.Errorf("empty user_id"))
				}

				// проверим, есть ли такой тип тр-ий
				_, ok := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
				if !ok {
					return utils.ErrInfo(fmt.Errorf("nonexistent type"))
				}

				p.TxMap = map[string][]byte{}

				// для статы
				p.TxIds = append(p.TxIds, string(p.TxSlice[1]))

				MethodName := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
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

				utils.WriteSelectiveLog("UPDATE transactions SET used=1 WHERE hex(hash) = " + string(utils.Md5(transactionBinaryDataFull)))
				affect, err := p.ExecSqlGetAffect("UPDATE transactions SET used=1 WHERE hex(hash) = ?", utils.Md5(transactionBinaryDataFull))
				if err != nil {
					utils.WriteSelectiveLog(err)
					return utils.ErrInfo(err)
				}
				utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))

				// даем юзеру понять, что его тр-ия попала в блок
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
