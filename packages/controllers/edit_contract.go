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
	"encoding/json"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"regexp"
)

type editContractPage struct {
	Alert               string
	Lang                map[string]string
	Data                map[string]string
	DataContractHistory []map[string]string
	WalletId            int64
	CitizenId           int64
	TxType              string
	TxTypeId            int64
	TxActivateType      string
	TxActivateTypeId    int64
	Confirm             bool
	TableName           string
	StateId             int64
	Global              string
}

func (c *Controller) EditContract() (string, error) {

	txType := "EditContract"
	txTypeId := utils.TypeInt(txType)

	global := c.r.FormValue("global")
	prefix := "global"
	if global == "" || global == "0" {
		prefix = c.StateIdStr
		global = "0"
	}

	id := utils.StrToInt64(c.r.FormValue("id"))
	name := c.r.FormValue("name")
	if id == 0 {
		// @ - global or alien state
		if len(name) > 0 && name[:1] == `@` {
			name = name[1:]
			r, _ := regexp.Compile(`([0-9]+)`)
			stateId := r.FindString(name)
			if len(stateId) > 0 {
				prefix = stateId
			}
			r, _ = regexp.Compile(`([\w]+)`)
			name = r.FindString(name)
		}
		if len(name) > 0 && !utils.CheckInputData_(name, "string", "") {
			return "", utils.ErrInfo("Incorrect name")
		}
	}

	var data map[string]string
	var dataContractHistory []map[string]string
	var rbId int64
	var err error
	var cont_wallet int64
	for i := 0; i < 10; i++ {
		if i == 0 {
			if id != 0 {
				data, err = c.OneRow(`SELECT * FROM "`+prefix+`_smart_contracts" WHERE id = ?`, id).String()
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			} else {
				data, err = c.OneRow(`SELECT * FROM "`+prefix+`_smart_contracts" WHERE name = ?`, name).String()
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			}
			if data[`wallet_id`] == `NULL` {
				data[`wallet`] = ``
			} else {
				cont_wallet = utils.StrToInt64(data[`wallet_id`])
				data[`wallet`] = lib.AddressToString(cont_wallet)
			}
			if data[`active`] == `NULL` {
				data[`active`] = ``
			}
			if len(data[`conditions`]) == 0 {
				data[`conditions`] = "ContractConditions(`MainCondition`)"
			}
			rbId = utils.StrToInt64(data["rb_id"])
		} else {
			data, err := c.OneRow(`SELECT data, block_id FROM "rollback" WHERE rb_id = ?`, rbId).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			var messageMap map[string]string
			json.Unmarshal([]byte(data["data"]), &messageMap)
			//			fmt.Printf("%s", messageMap)
			rbId = utils.StrToInt64(messageMap["rb_id"])
			messageMap["block_id"] = data["block_id"]
			dataContractHistory = append(dataContractHistory, messageMap)
		}
		if rbId == 0 {
			break
		}
	}
	TemplateStr, err := makeTemplate("edit_contract", "editContract", &editContractPage{
		Alert:               c.Alert,
		Lang:                c.Lang,
		WalletId:            c.SessWalletId,
		Data:                data,
		DataContractHistory: dataContractHistory,
		Global:              global,
		CitizenId:           c.SessCitizenId,
		TxType:              txType,
		TxTypeId:            txTypeId,
		Confirm:             c.SessWalletId == cont_wallet,
		TxActivateType:      `ActivateContract`,
		TxActivateTypeId:    utils.TypeInt(`ActivateContract`),
		StateId:             c.SessStateId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
