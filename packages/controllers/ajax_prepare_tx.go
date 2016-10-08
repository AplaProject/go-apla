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

package controllers

import (
	"fmt"
	"strings"

	"github.com/DayLightProject/go-daylight/packages/lib"
	"github.com/DayLightProject/go-daylight/packages/script"
	"github.com/DayLightProject/go-daylight/packages/smart"
)

const APrepareTx = `ajax_prepare_tx`

type PrepareTxJson struct {
	ForSign string `json:"forsign"`
	Time    uint32 `json:"time"`
	Error   string `json:"error"`
}

func init() {
	newPage(APrepareTx, `json`)
}

func (c *Controller) AjaxPrepareTx() interface{} {
	var (
		result PrepareTxJson
		err    error
	)
	cntname := c.r.FormValue(`TxName`)
	contract := smart.GetContract(cntname)
	if contract == nil || contract.Block.Info.(*script.ContractInfo).Tx == nil {
		err = fmt.Errorf(`there is not %s contract %v`, cntname, contract)
	} else {
		info := (*contract).Block.Info.(*script.ContractInfo)
		result.Time = lib.Time32()
		userId := c.SessWalletId
		if c.SessStateId > 0 {
			userId = c.SessCitizenId
		}
		forsign := fmt.Sprintf("%d,%d,%d,%d", info.Id /*+smart.CNTOFF*/, result.Time, userId, c.SessStateId)

		for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
			val := c.r.FormValue(fitem.Name)
			if len(val) == 0 && !strings.Contains(fitem.Tags, `optional`) {
				err = fmt.Errorf(`%s is empty`, fitem.Name)
				break
			}
			forsign += fmt.Sprintf(",%v", val)
		}
		result.ForSign = forsign
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
