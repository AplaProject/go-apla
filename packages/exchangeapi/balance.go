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

package exchangeapi

import (
	"net/http"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type Balance struct {
	Error  string `json:"error"`
	Amount string `json:"amount"`
	EGS    string `json:"egs"`
}

func balance(r *http.Request) interface{} {
	var result Balance

	wallet := lib.StringToAddress(r.FormValue(`wallet`))
	if wallet == 0 {
		result.Error = `Wallet is invalid`
		return result
	}
	total, err := utils.DB.Single(`SELECT amount FROM dlt_wallets WHERE wallet_id = ?`, wallet).String()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Amount = total
	result.EGS = lib.EGSMoney(total)
	return result
}
