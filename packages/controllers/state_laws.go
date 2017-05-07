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

type stateLawsPage struct {
	Alert              string
	SignData           string
	ShowSignData       bool
	Lang               map[string]string
	WalletId           int64
	CitizenId          int64
	TxType             string
	TxTypeId           int64
	TimeNow            int64
	AllStateParameters []string
	StateLaws          []map[string]string
}

func (c *Controller) StateLaws() (string, error) {

	var err error

	txType := "StateLaws"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	stateLaws, err := c.GetAll(`SELECT * FROM ea_state_laws`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	allStateParameters, err := c.GetList(`SELECT parameter FROM ea_state_parameters`).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("state_laws", "stateLaws", &stateLawsPage{
		Alert:              c.Alert,
		Lang:               c.Lang,
		ShowSignData:       c.ShowSignData,
		SignData:           "",
		WalletId:           c.SessWalletId,
		CitizenId:          c.SessCitizenId,
		StateLaws:          stateLaws,
		AllStateParameters: allStateParameters,
		TimeNow:            timeNow,
		TxType:             txType,
		TxTypeId:           txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
