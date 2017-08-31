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

package api

import (
	"fmt"
	"net/http"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

type balanceResult struct {
	Amount string `json:"amount"`
	EGS    string `json:"egs"`
}

func balance(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	wallet := converter.StringToAddress(data.params[`wallet`].(string))
	if wallet == 0 {
		logger.LogError(consts.RouteError, data.params[`wallet`].(string))
		return errorAPI(w, fmt.Sprintf(`Wallet %s is invalid`, data.params[`wallet`].(string)), http.StatusBadRequest)
	}
	total, err := model.Single(`SELECT amount FROM dlt_wallets WHERE wallet_id = ?`, wallet).String()
	if err != nil {
		logger.LogError(consts.DBError, err)
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = &balanceResult{Amount: total, EGS: converter.EGSMoney(total)}
	return nil
}
