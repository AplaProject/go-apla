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
	"encoding/json"
	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/shopspring/decimal"
)

const nWalletHistory = `wallet_history`

type walletPage struct {
	Data   *CommonPage
	Wallet string
	IsData bool
	List   []map[string]interface{}
}

func init() {
	newPage(nWalletHistory)
}

// WalletHistory is a controller for displaying wallet's history
func (c *Controller) WalletHistory() (string, error) {
	list := make([]map[string]interface{}, 0)
	walletID := converter.StringToAddress(c.r.FormValue("wallet"))
	if walletID == 0 {
		walletID = c.SessWalletID
	}
	wallet := &model.DltWallet{}
	err := wallet.GetWallet(walletID)
	if err != nil {
		return ``, utils.ErrInfo(err)
	}
	//current, err := c.OneRow(`select amount, rb_id from dlt_wallets where wallet_id=?`, walletID).String()

	rb := wallet.RollbackID
	if rb != 0 {
		balance, err := decimal.NewFromString(wallet.Amount)
		if err != nil {
			return ``, utils.ErrInfo(err)
		}
		for len(list) <= 100 && rb > 0 {
			var data map[string]string
			rollback := &model.Rollback{}
			err := rollback.Get(rb)
			if err != nil {
				return ``, utils.ErrInfo(err)
			}
			if err = json.Unmarshal([]byte(rollback.Data), &data); err != nil {
				return ``, utils.ErrInfo(err)
			}
			rb, err = strconv.ParseInt(data[`rb_id`], 10, 64)
			if err != nil {
				logger.LogInfo(consts.StrToIntError, data["rb_id"])
			}
			if amount, ok := data[`amount`]; ok {
				var dif decimal.Decimal
				val, _ := decimal.NewFromString(amount)
				if balance.Cmp(val) > 0 {
					dif = balance.Sub(val)
				} else {
					dif = val.Sub(balance)
				}
				list = append(list, map[string]interface{}{`block_id`: string(rollback.BlockID),
					`amount`:  converter.EGSMoney(dif.String()),
					`balance`: converter.EGSMoney(balance.String()), `inc`: balance.Cmp(val) > 0})
				balance = val
			}
		}
	}
	pageData := walletPage{Data: c.Data, List: list, IsData: len(list) > 0, Wallet: converter.AddressToString(walletID)}
	return proceedTemplate(c, nWalletHistory, &pageData)
}
