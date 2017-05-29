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
	WalletID            int64
	CitizenID           int64
	TxType              string
	TxTypeID            int64
	TxActivateType      string
	TxActivateTypeID    int64
	Confirm             bool
	TableName           string
	StateID             int64
	Global              string
}

// EditContract is a handler function for editing contracts
func (c *Controller) EditContract() (string, error) {

	txType := "EditContract"
	txTypeID := utils.TypeInt(txType)

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
			stateID := r.FindString(name)
			if len(stateID) > 0 {
				prefix = stateID
			}
			r, _ = regexp.Compile(`([\w]+)`)
			name = r.FindString(name)
		}
		if len(name) > 0 && !utils.CheckInputData(name, "string") {
			return "", utils.ErrInfo("Incorrect name")
		}
	}

	var data map[string]string
	var dataContractHistory []map[string]string
	var rbID int64
	var err error
	var contWallet int64
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
				contWallet = utils.StrToInt64(data[`wallet_id`])
				data[`wallet`] = lib.AddressToString(contWallet)
			}
			if data[`active`] == `NULL` {
				data[`active`] = ``
			}
			if len(data[`conditions`]) == 0 {
				data[`conditions`] = "ContractConditions(`MainCondition`)"
			}
			rbID = utils.StrToInt64(data["rb_id"])
		} else {
			data, err := c.OneRow(`SELECT data, block_id FROM "rollback" WHERE rb_id = ?`, rbID).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			var messageMap map[string]string
			json.Unmarshal([]byte(data["data"]), &messageMap)
			//			fmt.Printf("%s", messageMap)
			rbID = utils.StrToInt64(messageMap["rb_id"])
			messageMap["block_id"] = data["block_id"]
			dataContractHistory = append(dataContractHistory, messageMap)
		}
		if rbID == 0 {
			break
		}
	}
	TemplateStr, err := makeTemplate("edit_contract", "editContract", &editContractPage{
		Alert:               c.Alert,
		Lang:                c.Lang,
		WalletID:            c.SessWalletID,
		Data:                data,
		DataContractHistory: dataContractHistory,
		Global:              global,
		CitizenID:           c.SessCitizenID,
		TxType:              txType,
		TxTypeID:            txTypeID,
		Confirm:             c.SessWalletID == contWallet,
		TxActivateType:      `ActivateContract`,
		TxActivateTypeID:    utils.TypeInt(`ActivateContract`),
		StateID:             c.SessStateID})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
