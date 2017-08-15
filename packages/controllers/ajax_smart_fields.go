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
	"strconv"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/shopspring/decimal"
)

const aSmartFields = `ajax_smart_fields`

// SmartFieldsJSON is a structure for the answer of ajax_smart_fields ajax request
type SmartFieldsJSON struct {
	Fields   string `json:"fields"`
	Price    int64  `json:"price"`
	Valid    bool   `json:"valid"`
	Approved int64  `json:"approved"`
	Error    string `json:"error"`
}

func init() {
	newPage(aSmartFields, `json`)
}

// AjaxSmartFields is a controller of ajax_smart_fields request
func (c *Controller) AjaxSmartFields() interface{} {
	var (
		result SmartFieldsJSON
		err    error
	)
	stateID := converter.StrToInt64(c.r.FormValue(`state_id`))
	stateStr := converter.Int64ToStr(stateID)
	if !model.IsTable(stateStr+`_citizens`) || !model.IsTable(stateStr+`_citizenship_requests`) {
		result.Error = `Basic app is not installed`
		return result
	}

	citizen := &model.Citizen{ID: c.SessWalletID}
	citizen.SetTablePrefix(stateStr)
	if exist, err := citizen.IsExists(); err != nil {
		result.Error = err.Error()
		return result
	} else if exist == true {
		result.Approved = 2
		return result
	}

	request := &model.CitizenshipRequest{}
	request.SetTablePrefix(stateStr)
	err = request.GetByWalletOrdered(c.SessWalletID)
	if err == nil {
		if request.ID > 0 {
			result.Approved = request.Approved
		} else {
			cntname := c.r.FormValue(`contract_name`)
			contract := smart.GetContract(cntname, uint32(stateID))
			if contract == nil || contract.Block.Info.(*script.ContractInfo).Tx == nil {
				err = fmt.Errorf(`there is not %s contract`, cntname)
			} else {
				fields := make([]string, 0)
			main:
				for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
					if strings.Index(fitem.Tags, `hidden`) >= 0 {
						continue
					}
					for _, tag := range []string{`date`, `polymap`, `map`, `image`} {
						if strings.Index(fitem.Tags, tag) >= 0 {
							fields = append(fields, fmt.Sprintf(`{"name":"%s", "htmlType":"%s", "txType":"%s", "title":"%s"}`,
								fitem.Name, tag, fitem.Type.String(), fitem.Name))
							continue main
						}
					}
					if fitem.Type.String() == `string` || fitem.Type.String() == `int64` || fitem.Type.String() == script.Decimal {
						fields = append(fields, fmt.Sprintf(`{"name":"%s", "htmlType":"textinput", "txType":"%s", "title":"%s"}`,
							fitem.Name, fitem.Type.String(), fitem.Name))
					}
				}
				result.Fields = fmt.Sprintf(`[%s]`, strings.Join(fields, `,`))

				if err == nil {
					stateParameter := &model.StateParameter{}
					stateParameter.SetTablePrefix(stateStr)
					err := stateParameter.GetByName("citizenship_price")
					if err == nil {
						result.Price, _ = strconv.ParseInt(stateParameter.Value, 10, 64)
						dltWallet := &model.DltWallet{}
						err = dltWallet.GetWallet(c.SessWalletID)
						dPrice, _ := decimal.NewFromString(stateParameter.Value)
						wltAmount, err := decimal.NewFromString(dltWallet.Amount)
						if err == nil {
							result.Valid = (err == nil && wltAmount.Cmp(dPrice) >= 0)
						}
					}
				}

			}
		}
	}
	//	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
