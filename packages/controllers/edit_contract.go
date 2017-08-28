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
	"regexp"
	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
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
		prefix = c.StateIDStr
		global = "0"
	}

	id, err := strconv.ParseInt(c.r.FormValue("id"), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, c.r.FormValue("id"))
	}
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
	var contWallet int64
	for i := 0; i < 10; i++ {
		if i == 0 {
			smartContract := &model.SmartContract{}
			smartContract.SetTablePrefix(prefix)
			if id != 0 {
				err = smartContract.GetByID(id)
			} else {
				err = smartContract.GetByName(name)
			}
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			data := smartContract.ToMap()
			data[`wallet`] = converter.AddressToString(smartContract.WalletID)
			if len(smartContract.Conditions) == 0 {
				data[`conditions`] = "ContractConditions(`MainCondition`)"
			}
			rbID = smartContract.RbID
		} else {
			rollback := &model.Rollback{}
			err = rollback.Get(rbID)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			data := rollback.ToMap()
			var messageMap map[string]string
			json.Unmarshal([]byte(data["data"]), &messageMap)
			rbID, err = strconv.ParseInt(messageMap["rb_id"], 10, 64)
			if err != nil {
				logger.LogInfo(consts.StrtoInt64Error, messageMap["rb_id"])
			}
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
