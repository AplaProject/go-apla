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
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type editStateParametersPage struct {
	Alert              string
	Lang               map[string]string
	WalletID           int64
	CitizenID          int64
	StateID            int64
	TxType             string
	TxTypeID           int64
	TimeNow            int64
	StateParameters    map[string]string
	AllStateParameters []string
}

// EditStateParameters is a handle function for changing state parameters
func (c *Controller) EditStateParameters() (string, error) {

	var err error

	txType := "EditStateParameters"
	timeNow := time.Now().Unix()

	name := c.r.FormValue(`name`)

	stateParameters, err := c.OneRow(`SELECT * FROM "`+c.StateIDStr+`_state_parameters" WHERE name = ?`, name).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	allStateParameters, err := c.GetList(`SELECT name FROM "` + c.StateIDStr + `_state_parameters"`).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("edit_state_parameters", "editStateParameters", &editStateParametersPage{
		Alert:              c.Alert,
		Lang:               c.Lang,
		WalletID:           c.SessWalletID,
		CitizenID:          c.SessCitizenID,
		StateID:            c.StateID,
		StateParameters:    stateParameters,
		AllStateParameters: allStateParameters,
		TimeNow:            timeNow,
		TxType:             txType,
		TxTypeID:           utils.TypeInt(txType)})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
