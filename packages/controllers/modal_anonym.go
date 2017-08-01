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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type modalAnonymPage struct {
	Lang         map[string]string
	Title        string
	CountSign    int
	MyWalletData map[string]string
	Address      string
	WalletID     int64
	CitizenID    int64
}

// ModalAnonym shows QR code of the wallet
func (c *Controller) ModalAnonym() (string, error) {
	wallet := &model.DltWallet{}
	err := wallet.GetWallet(c.SessWalletID)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	data := wallet.ToMap()
	data[`address`] = converter.AddressToString(c.SessWalletID)
	log.Debug("MyWalletData %v", data)

	TemplateStr, err := makeTemplate("modal_anonym", "modalAnonym", &modalAnonymPage{
		Lang:         c.Lang,
		MyWalletData: data,
		Title:        "modalAnonym",
		WalletID:     c.SessWalletID,
		CitizenID:    c.SessCitizenID,
		Address:      c.SessAddress,
	})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
