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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func (p *Parser) ParseDataLite() error {
	p.dataPre()
	if p.dataType != 0 { // парсим только блоки
		return utils.ErrInfo(fmt.Errorf("incorrect dataType"))
	}

	err := p.ParseBlock()
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
