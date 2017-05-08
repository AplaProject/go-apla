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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type newStatePage struct {
	Alert     string
	Lang      map[string]string
	WalletId  int64
	CitizenId int64
	TxType    string
	TxTypeId  int64
	TimeNow   int64
}

func (c *Controller) NewState() (string, error) {

	var err error

	txType := "NewState"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	TemplateStr, err := makeTemplate("new_state", "newState", &newStatePage{
		Alert:     c.Alert,
		Lang:      c.Lang,
		WalletId:  c.SessWalletId,
		CitizenId: c.SessCitizenId,
		TimeNow:   timeNow,
		TxType:    txType,
		TxTypeId:  txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
