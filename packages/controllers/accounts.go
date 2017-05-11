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

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const NAccounts = `accounts`

type AccountInfo struct {
	AccountId int64  `json:"account_id"`
	Address   string `json:"address"`
	Amount    string `json:"amount"`
}

type accountsPage struct {
	Data     *CommonPage
	List     []AccountInfo
	Currency string
	TxType   string
	TxTypeId int64
	Unique   string
}

func init() {
	newPage(NAccounts)
}

func (c *Controller) Accounts() (string, error) {

	data := make([]AccountInfo, 0)

	cents, _ := utils.StateParam(c.SessStateId, `money_digit`)
	digit := utils.StrToInt(cents)

	currency, _ := utils.StateParam(c.SessStateId, `currency_name`)

	newAccount := func(account int64, amount string) {
		if amount == `NULL` {
			amount = ``
		} else if len(amount) > 0 {
			if digit > 0 {
				if len(amount) < digit+1 {
					amount = strings.Repeat(`0`, digit+1-len(amount)) + amount
				}
				amount = amount[:len(amount)-digit] + `.` + amount[len(amount)-digit:]
			}
		}
		data = append(data, AccountInfo{AccountId: account, Address: lib.AddressToString(account),
			Amount: amount})
	}

	amount, err := c.Single(fmt.Sprintf(`select amount from "%d_accounts" where citizen_id=?`,
		c.SessStateId), c.SessCitizenId).String()
	if err != nil {
		return ``, err
	}
	if len(amount) > 0 {
		newAccount(c.SessCitizenId, amount)
	} else {
		newAccount(c.SessCitizenId, `NULL`)
	}

	list, err := c.GetAll(fmt.Sprintf(`select anon.*, acc.amount from "%d_anonyms" as anon
	left join "%[1]d_accounts" as acc on acc.citizen_id=anon.id_anonym
	where anon.id_citizen=?`, c.SessStateId), -1, c.SessCitizenId)
	if err != nil {
		return ``, err
	}

	for _, item := range list {
		newAccount(utils.StrToInt64(item[`id_anonym`]), item[`amount`])
	}
	txType := "NewAccount"
	pageData := accountsPage{Data: c.Data, List: data, Currency: currency, TxType: txType,
		TxTypeId: utils.TypeInt(txType), Unique: ``}
	return proceedTemplate(c, NAccounts, &pageData)
}
