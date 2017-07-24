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
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
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

	stateParameters := &model.StateParameters{}
	if err := stateParameters.GetByName("money_digit"); err != nil {
		return ``, err
	}
	digit := converter.StrToInt(stateParameters.Value)

	if err := stateParameters.GetByName("currency_name"); err != nil {
		return ``, err
	}
	currency := stateParameters.Value

	newAccount := func(account int64, amount decimal.Decimal) {
		strAmount := formatAmount(digit, amount)
		data = append(data, AccountInfo{AccountID: account, Address: converter.AddressToString(account),
			Amount: strAmount})
	}

	account := &model.Accounts{}
	account.SetTablePrefix(c.SessCitizenID)
	err := account.Get(c.SessStateID)
	if err != nil {
		return ``, err
	}
	newAccount(c.SessCitizenID, account.Amount)

	list, err := c.GetAll(fmt.Sprintf(`select anon.*, acc.amount from "%d_anonyms" as anon
	left join "%[1]d_accounts" as acc on acc.citizen_id=anon.id_anonym
	where anon.id_citizen=?`, c.SessStateID), -1, c.SessCitizenID)
	if err != nil {
		return ``, err
	}

	for _, item := range list {
		amount, _ := decimal.NewFromString(item[`amount`])
		newAccount(converter.StrToInt64(item[`id_anonym`]), amount)
	}
	txType := "NewAccount"
	pageData := accountsPage{Data: c.Data, List: data, Currency: currency, TxType: txType,
		TxTypeID: utils.TypeInt(txType), Unique: ``}
	return proceedTemplate(c, nAccounts, &pageData)
}
