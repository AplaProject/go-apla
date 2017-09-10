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

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

// Balance is the result structure of balamce handler
type Balance struct {
	Error  string `json:"error"`
	Amount string `json:"amount"`
	EGS    string `json:"egs"`
}

func balance(r *http.Request) interface{} {
	logger.LogDebug(consts.FuncStarted, "")
	var result Balance

	wallet := converter.StringToAddress(r.FormValue(`wallet`))
	if wallet == 0 {
		result.Error = `Wallet is invalid`
		return result
	}
	total, err := model.Single(`SELECT amount FROM dlt_wallets WHERE wallet_id = ?`, wallet).String()
	if err != nil {
		logger.LogError(consts.DBError, err)
		result.Error = err.Error()
		return result
	}
	result.Amount = total
	result.EGS = converter.EGSMoney(total)
	return result
}
