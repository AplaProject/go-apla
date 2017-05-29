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
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type forgingPage struct {
	Lang         map[string]string
	Title        string
	MyWalletData map[string]string
	WalletID     int64
	CitizenID    int64
	TxType       string
	TxTypeID     int64
	TimeNow      int64
}

// Forging is a controller for DLTChangeHostVote transaction
func (c *Controller) Forging() (string, error) {

	txType := "DLTChangeHostVote"
	timeNow := utils.Time()

	MyWalletData, err := c.OneRow("SELECT host, address_vote, fuel_rate FROM dlt_wallets WHERE wallet_id = ?", c.SessWalletID).String()
	MyWalletData[`address`] = lib.AddressToString(c.SessWalletID)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("MyWalletData %v", MyWalletData)

	TemplateStr, err := makeTemplate("forging", "forging", &forgingPage{
		Lang:         c.Lang,
		MyWalletData: MyWalletData,
		Title:        "modalAnonym",
		WalletID:     c.SessWalletID,
		CitizenID:    c.SessCitizenID,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeID:     utils.TypeInt(txType)})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
