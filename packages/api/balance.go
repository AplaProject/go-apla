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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type balanceResult struct {
	Amount string `json:"amount"`
	Money  string `json:"money"`
}

func balance(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	ecosystemId, _, err := checkEcosystem(w, data, logger)
	if err != nil {
		return err
	}
	keyID := converter.StringToAddress(data.params[`wallet`].(string))
	if keyID == 0 {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "value": data.params["wallet"].(string)}).Error("converting wallet to address")
		return errorAPI(w, `E_INVALIDWALLET`, http.StatusBadRequest, data.params[`wallet`].(string))
	}

	key := &model.Key{}
	key.SetTablePrefix(ecosystemId)
	_, err = key.Get(keyID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting Key for wallet")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	data.result = &balanceResult{Amount: key.Amount, Money: converter.EGSMoney(key.Amount)}
	return nil
}
