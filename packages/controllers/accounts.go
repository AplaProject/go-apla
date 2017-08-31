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
	"strconv"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/shopspring/decimal"
)

const nAccounts = `accounts`

// AccountInfo is a structure for the list of the accounts
type AccountInfo struct {
	AccountID int64  `json:"account_id"`
	Address   string `json:"address"`
	Amount    string `json:"amount"`
}

type accountsPage struct {
	Data     *CommonPage
	List     []AccountInfo
	Currency string
	TxType   string
	TxTypeID int64
	Unique   string
}

func init() {
	newPage(nAccounts)
}

func formatAmount(digit int, amount decimal.Decimal) string {
	amountStr := amount.String()
	if digit > 0 {
		if len(amountStr) < digit+1 {
			amountStr = strings.Repeat(`0`, digit+1-len(amountStr)) + amountStr
		}
		amountStr = amountStr[:len(amountStr)-digit] + `.` + amountStr[len(amountStr)-digit:]
	}
	return amountStr
}

// Accounts is a controller for accounts page
func (c *Controller) Accounts() (string, error) {

	data := make([]AccountInfo, 0)

	stateParameter := &model.StateParameter{}
	if err := stateParameter.GetByName("money_digit"); err != nil {
		return ``, err
	}
	digit, err := strconv.Atoi(stateParameter.Value)
	if err != nil {
		logger.LogInfo(consts.StrToIntError, stateParameter.Value)
	}

	if err := stateParameter.GetByName("currency_name"); err != nil {
		return ``, err
	}
	currency := stateParameter.Value

	newAccount := func(account int64, amount decimal.Decimal) {
		strAmount := formatAmount(digit, amount)
		data = append(data, AccountInfo{AccountID: account, Address: converter.AddressToString(account),
			Amount: strAmount})
	}

	account := &model.Account{}
	account.SetTablePrefix(c.SessCitizenID)
	err = account.Get(c.SessStateID)
	if err != nil {
		return ``, err
	}
	newAccount(c.SessCitizenID, account.Amount)

	aa := &model.AnonAmount{}
	amounts, err := aa.Get(c.SessStateID, c.SessCitizenID)
	if err != nil {
		return ``, err
	}

	for _, item := range amounts {
		newAccount(item.IDAnonym, item.Amount)
	}
	txType := "NewAccount"
	pageData := accountsPage{Data: c.Data, List: data, Currency: currency, TxType: txType,
		TxTypeID: utils.TypeInt(txType), Unique: ``}
	return proceedTemplate(c, nAccounts, &pageData)
}
