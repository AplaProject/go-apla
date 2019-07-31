// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package api

import (
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type balanceResult struct {
	Amount string `json:"amount"`
	Money  string `json:"money"`
}

func (m Mode) getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	form := &ecosystemForm{
		Validator: m.EcosysIDValidator,
	}

	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)

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
