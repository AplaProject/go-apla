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

type restoreAccessPage struct {
	Alert              string
	Active             int64
	Request            int64
	Lang               map[string]string
	WalletID           int64
	CitizenID          int64
	TxType             string
	TxTypeID           int64
	TimeNow            int64
	StateSmartLaws     map[string]string
	AllStateParameters []string
}

// RestoreAccess is a restore access control
func (c *Controller) RestoreAccess() (string, error) {

	txType := "RestoreAccessActive"
	timeNow := time.Now().Unix()

	data, err := c.OneRow("SELECT active FROM system_restore_access WHERE state_id  =  ?", c.SessStateID).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	active := data["active"]

	var request int64
	if data["time"] > 0 {
		request = data["time"]
	}

	TemplateStr, err := makeTemplate("restore_access", "restoreAccess", &restoreAccessPage{
		Alert:     c.Alert,
		Lang:      c.Lang,
		Active:    active,
		Request:   request,
		TimeNow:   timeNow,
		TxType:    txType,
		TxTypeID:  utils.TypeInt(txType),
		WalletID:  c.SessWalletID,
		CitizenID: c.SessCitizenID,
	})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
