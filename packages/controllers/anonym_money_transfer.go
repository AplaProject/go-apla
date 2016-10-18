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

type anonymMoneyTransferPage struct {
	Lang         map[string]string
	Title        string
	CountSign    int
	CountSignArr []int
	SignData     string
	ShowSignData bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	WalletId     int64
	CitizenId    int64
	Commission   int64
}

func (c *Controller) AnonymMoneyTransfer() (string, error) {

	txType := "DLTTransfer"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	fPrice, err := c.Single(`SELECT value->'dlt_transfer' FROM system_parameters WHERE name = ?`, "op_price").Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	fuelRate, err := c.Single(`SELECT value FROM system_parameters WHERE name = ?`, "fuel_rate").Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	commission := int64(fPrice * fuelRate)

	log.Debug("sessCitizenId %d SessWalletId %d SessStateId %d", c.SessCitizenId, c.SessWalletId, c.SessStateId)

	TemplateStr, err := makeTemplate("anonym_money_transfer", "anonymMoneyTransfer", &anonymMoneyTransferPage{
		CountSignArr: c.CountSignArr,
		CountSign:    c.CountSign,
		Lang:         c.Lang,
		Title:        "anonymMoneyTransfer",
		ShowSignData: c.ShowSignData,
		SignData:     "",
		WalletId:     c.SessWalletId,
		CitizenId:    c.SessCitizenId,
		Commission : commission,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
