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
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type stateTablePage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Lang         map[string]string
	WalletId  int64
	CitizenId int64
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	Tables []map[string]string
}

func (c *Controller) StateTable() (string, error) {

	var err error

	txType := "stateTable"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	stateId := utils.StrToInt64(c.Parameters["stateId"])

	tables, err := c.GetAll(`SELECT * FROM `+utils.Int64ToStr(stateId)+`_tables`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("state_list", "stateTable", &stateTablePage {
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		WalletId: c.SessWalletId,
		CitizenId: c.SessCitizenId,
		CountSignArr: c.CountSignArr,
		Tables : tables,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
