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
	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type rowHistoryPage struct {
	Alert     string
	Lang      map[string]string
	History   []map[string]string
	WalletID  int64
	CitizenID int64
	TableName string
	StateID   int64
	Global    string
	Columns   map[string]string
}

// RowHistory returns rollback data of the table
func (c *Controller) RowHistory() (string, error) {

	var history []map[string]string
	rbID, err := strconv.ParseInt(c.r.FormValue("rbId"), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrToIntError, c.r.FormValue("rbId"))
	}
	if rbID < 1 {
		return "", utils.ErrInfo(`Incorrect rbId`)
	}
	var tableName string
	if utils.CheckInputData(c.r.FormValue("tableName"), "string") {
		tableName = c.r.FormValue("tableName")
	}

	global := c.r.FormValue("global")
	prefix := c.StateIDStr
	if global == "1" {
		prefix = "global"
	} else {
		global = "0"
	}
	t := &model.Table{}
	t.SetTablePrefix(prefix)
	columns, err := t.GetPermissions(tableName, "update")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	columns["id"] = ""
	columns["block_id"] = ""
	for i := 0; i < 100; i++ {
		rollback := &model.Rollback{}
		err := rollback.Get(rbID)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		var messageMap map[string]string
		json.Unmarshal([]byte(rollback.Data), &messageMap)
		rbID, err = strconv.ParseInt(messageMap["rb_id"], 10, 64)
		if err != nil {
			logger.LogInfo(consts.StrToIntError, messageMap["rb_id"])
		}
		messageMap["block_id"] = string(rollback.BlockID)
		history = append(history, messageMap)
		if rbID == 0 {
			break
		}
	}

	TemplateStr, err := makeTemplate("row_history", "rowHistory", &rowHistoryPage{
		Alert:     c.Alert,
		Lang:      c.Lang,
		WalletID:  c.SessWalletID,
		History:   history,
		CitizenID: c.SessCitizenID,
		TableName: tableName,
		Global:    global,
		Columns:   columns,
		StateID:   c.SessStateID})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
