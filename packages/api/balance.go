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
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type balanceResult struct {
	Amount string `json:"amount"`
	Money  string `json:"money"`
}

func getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	form := &ecosystemForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	logger := getLogger(r)

	keyID := converter.StringToAddress(params["wallet"])
	if keyID == 0 {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "value": params["wallet"]}).Error("converting wallet to address")
		errorResponse(w, errInvalidWallet.Errorf(params["wallet"]))
		return
	}

	key := &model.Key{}
	key.SetTablePrefix(form.EcosystemID)
	_, err := key.Get(keyID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting Key for wallet")
		errorResponse(w, err)
		return
	}

	jsonResponse(w, &balanceResult{
		Amount: key.Amount,
		Money:  converter.EGSMoney(key.Amount),
	})
}
