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

type restoreAccessPage struct {
	Alert              string
	Active             int64
	Request            int32
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

	sra := &model.SystemRestoreAccess{}
	err := sra.Get(c.SessStateID)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	var request int32
	if sra.Time > 0 {
		request = int32(sra.Time)
	}

	TemplateStr, err := makeTemplate("restore_access", "restoreAccess", &restoreAccessPage{
		Alert:     c.Alert,
		Lang:      c.Lang,
		Active:    sra.Active,
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
