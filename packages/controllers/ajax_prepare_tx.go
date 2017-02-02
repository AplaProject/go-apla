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

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"regexp"
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

func (c *Controller) checkTx() (contract *smart.Contract, err error) {
	cntname := c.r.FormValue(`TxName`)
	contract = smart.GetContract(cntname, uint32(c.SessStateId))
	if contract == nil /*|| contract.Block.Info.(*script.ContractInfo).Tx == nil*/ {
		err = fmt.Errorf(`there is not %s contract %v`, cntname, contract)
	} else if contract.Block.Info.(*script.ContractInfo).Tx != nil {
		for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
			if strings.Index(fitem.Tags, `image`) >= 0 {
				continue
			}
			val := strings.TrimSpace(c.r.FormValue(fitem.Name))
			if len(val) == 0 && !strings.Contains(fitem.Tags, `optional`) {
				err = fmt.Errorf(`%s is empty`, fitem.Name)
				break
			}
			if strings.Index(fitem.Tags, `address`) >= 0 {
				addr := lib.StringToAddress(val)
				if addr == 0 {
					err = fmt.Errorf(`Address %s is not valid`, val)
					break
				}
			}
			if fitem.Type.String() == `decimal.Decimal` {
				re := regexp.MustCompile(`^\d+$`) //`^\d+\.?\d+?$`
				if !re.Match([]byte(val)) {
					err = fmt.Errorf(`The value of money %s is not valid`, val)
					break
				}
			}
		}
	}
	return
}

func (c *Controller) AjaxPrepareTx() interface{} {
	var (
		result PrepareTxJson
	)
	contract, err := c.checkTx()
	if err == nil {
		var flags uint8
		var isPublic []byte
		info := (*contract).Block.Info.(*script.ContractInfo)
		result.Time = lib.Time32()
		userId := uint64(c.SessWalletId)
		isPublic, err = c.Single(`select public_key_0 from dlt_wallets where wallet_id=?`, c.SessWalletId).Bytes()
		if err == nil && len(isPublic) == 0 {
			flags |= consts.TxfPublic
		}
		fmt.Println(`Prepare`, c.SessWalletId, c.SessCitizenId, c.SessStateId)
		/*		if c.SessStateId > 0 {
				userId = c.SessCitizenId
			}*/
		forsign := fmt.Sprintf("%d,%d,%d,%d,%d", info.Id, result.Time, userId, c.SessStateId, flags)
		if (*contract).Block.Info.(*script.ContractInfo).Tx != nil {
			for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
				if strings.Index(fitem.Tags, `image`) >= 0 {
					continue
				}
				val := strings.TrimSpace(c.r.FormValue(fitem.Name))
				if strings.Index(fitem.Tags, `address`) >= 0 {
					val = utils.Int64ToStr(lib.StringToAddress(val))
				} else if fitem.Type.String() == `decimal.Decimal` {
					val = strings.TrimLeft(val, `0`)
				}
				forsign += fmt.Sprintf(",%v", val)
			}
		}
		result.ForSign = forsign
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
