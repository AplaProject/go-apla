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

type changeNodeKeyPage struct {
	Alert     string
	Lang      map[string]string
	WalletID  int64
	CitizenID int64
	NoPublic  bool
	TxType    string
	TxTypeID  int64
	TimeNow   int64
}

func (c *Controller) ChangeNodeKey() (string, error) {

	var err error

	txType := "ChangeNodeKeyDLT"
	txTypeID := utils.TypeInt(txType)
	timeNow := utils.Time()

	public, err := c.OneRow("SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?", c.SessWalletId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("change_node_key", "changeNodeKey", &changeNodeKeyPage{
		Alert:     c.Alert,
		Lang:      c.Lang,
		WalletID:  c.SessWalletId,
		CitizenID: c.SessCitizenId,
		TimeNow:   timeNow,
		TxType:    txType,
		NoPublic:  len(public) == 0,
		TxTypeID:  txTypeID})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
