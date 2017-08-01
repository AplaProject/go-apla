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
	"fmt"
	"time"

	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/shopspring/decimal"
)

type anonymMoneyTransferPage struct {
	Lang       map[string]string
	Title      string
	TxType     string
	TxTypeID   int64
	TimeNow    int64
	WalletID   int64
	CitizenID  int64
	Commission string
	Amount     string
}

// AnonymMoneyTransfer is a controller of the money transfer template page
func (c *Controller) AnonymMoneyTransfer() (string, error) {

	txType := "DLTTransfer"
	txTypeID := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	systemParameters := &model.SystemParameter{}
	value, err := systemParameters.GetJSONField(`value->'dlt_transfer'`, "op_price")
	fPrice, _ := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	err = systemParameters.Get("fuel_rate")
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	fuelRate, _ := decimal.NewFromString(systemParameters.Value)
	if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
		return ``, fmt.Errorf(`fuel rate must be greater than 0`)
	}

	commission := decimal.New(fPrice, 0).Mul(fuelRate)

	log.Debug("sessCitizenID %d SessWalletID %d SessStateID %d", c.SessCitizenID, c.SessWalletID, c.SessStateID)
	dltWallet := &model.DltWallet{}
	err = dltWallet.GetWallet(c.SessWalletID)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	TemplateStr, err := makeTemplate("anonym_money_transfer", "anonymMoneyTransfer", &anonymMoneyTransferPage{
		Lang:       c.Lang,
		Title:      "anonymMoneyTransfer",
		Amount:     dltWallet.Amount.String(),
		WalletID:   c.SessWalletID,
		CitizenID:  c.SessCitizenID,
		Commission: commission.String(),
		TimeNow:    timeNow,
		TxType:     txType,
		TxTypeID:   txTypeID})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
