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
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"strings"
)

const nEditWallet = `edit_wallet`

type editWalletPage struct {
	Alert    string
	Data     *CommonPage
	TxType   string
	TxTypeID int64
	//	Lang                map[string]string
	Info    map[string]string
	Unique  string
	StateID int64
}

func init() {
	newPage(nEditWallet)
}

// EditWallet is a controller for editing state's wallets
func (c *Controller) EditWallet() (string, error) {

	var (
		err   error
		data  map[string]string
		alert string
	)

	txType := "EditWallet"

	idaddr := lib.StripTags(c.r.FormValue("id"))
	var id int64
	if len(idaddr) > 0 {
		if idaddr[0] == '-' {
			id = utils.StrToInt64(idaddr)
		} else if strings.IndexByte(idaddr, '-') < 0 {
			id = int64(utils.StrToUint64(idaddr))
		} else {
			id = lib.StringToAddress(idaddr)
		}
		if id == 0 {
			alert = fmt.Sprintf(`Address %s is not valid.`, idaddr)
		}
	} else {
		id = c.SessWalletID
	}
	if id != 0 {
		data, err = c.OneRow(`SELECT * FROM "dlt_wallets" WHERE wallet_id = ?`, id).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(data) == 0 {
			alert = fmt.Sprintf(`Wallet %s [%d] has not been found.`, idaddr, id)
		} else {
			ret := data[`amount`]
			if ret != `0` {
				if len(ret) < consts.EGS_DIGIT+1 {
					ret = strings.Repeat(`0`, consts.EGS_DIGIT+1-len(ret)) + ret
				}
				ret = ret[:len(ret)-consts.EGS_DIGIT] + `.` + ret[len(ret)-consts.EGS_DIGIT:]
				data[`amount`] = ret
			}
			data[`address`] = lib.AddressToString(id)
			if data[`spending_contract`] == `NULL` {
				data[`spending_contract`] = ``
			}
			if data[`conditions_change`] == `NULL` {
				data[`conditions_change`] = ``
			}
		}
	}
	pageData := editWalletPage{Data: c.Data, StateID: c.SessStateID,
		Alert: alert, TxType: txType, TxTypeID: utils.TypeInt(txType), Info: data, Unique: ``}
	return proceedTemplate(c, nEditWallet, &pageData)
}
