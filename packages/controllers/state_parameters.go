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

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type stateParametersPage struct {
	Alert           string
	Lang            map[string]string
	WalletID        int64
	CitizenID       int64
	TxType          string
	TxTypeID        int64
	TimeNow         int64
	StateParameters []map[string]string
}

// StateParameters is a controller for the list of state parameters
func (c *Controller) StateParameters() (string, error) {
	txType := "StateParameters"
	timeNow := time.Now().Unix()

	stateParams := &model.StateParameters{}
	rows, err := stateParams.GetAllStateParameters(c.StateIDStr)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	stateParameters := make([]map[string]string, 0)
	for _, row := range rows {
		stateParameters = append(stateParameters, row.ToMap())
	}

	TemplateStr, err := makeTemplate("state_parameters", "stateParameters", &stateParametersPage{
		Alert:           c.Alert,
		Lang:            c.Lang,
		WalletID:        c.SessWalletID,
		CitizenID:       c.SessCitizenID,
		StateParameters: stateParameters,
		TimeNow:         timeNow,
		TxType:          txType,
		TxTypeID:        utils.TypeInt(txType)})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
