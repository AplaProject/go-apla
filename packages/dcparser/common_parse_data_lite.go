package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ParseDataLite() error {
	p.dataPre()
	if p.dataType != 0 { // парсим только блоки
		return utils.ErrInfo(fmt.Errorf("incorrect dataType"))
	}
	var err error
	p.Variables, err = p.GetAllVariables()
	if err != nil {
		return utils.ErrInfo(err)
	}
	err = p.ParseBlock()
	if err != nil {
		return utils.ErrInfo(err)
	}

	if len(p.BinaryData) > 0 {
		i := 0
		for {
			transactionSize := utils.DecodeLength(&p.BinaryData)
			if len(p.BinaryData) == 0 {
				return utils.ErrInfo(fmt.Errorf("empty BinaryData"))
			}
			// отчекрыжим одну транзакцию от списка транзакций
			transactionBinaryData := utils.BytesShift(&p.BinaryData, transactionSize)
			transactionBinaryDataFull := transactionBinaryData

			p.TxHash = string(utils.Md5(transactionBinaryData))
			p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)

			MethodName := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
			log.Debug("MethodName", MethodName+"Init")
			err_ := utils.CallMethod(p, MethodName+"Init")
			if _, ok := err_.(error); ok {
				log.Debug("%v", err)
				return utils.ErrInfo(err_.(error))
			}
			p.TxMap["md5hash"] = utils.Md5(transactionBinaryDataFull)
			p.TxMapArr = append(p.TxMapArr, p.TxMap)
			//p.TxMapsArr = append(p.TxMapsArr, p.TxMaps)
			if len(p.BinaryData) == 0 {
				break
			}
			i++
		}
	}

	return nil
}
