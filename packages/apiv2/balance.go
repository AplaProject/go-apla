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

package apiv2

import (
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
)

type balanceResult struct {
	Amount string `json:"amount"`
	Money  string `json:"money"`
}

func balance(w http.ResponseWriter, r *http.Request, data *apiData) error {

	ecosystemId, err := checkEcosystem(w, data)
	if err != nil {
		return err
	}
	keyID := converter.StringToAddress(data.params[`wallet`].(string))
	if keyID == 0 {
		return errorAPI(w, `E_INVALIDWALLET`, http.StatusBadRequest, data.params[`wallet`].(string))
	}

	key := &model.Key{}
	key.SetTablePrefix(ecosystemId)
	err = key.Get(keyID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	data.result = &balanceResult{Amount: key.Amount, Money: converter.EGSMoney(key.Amount)}
	return nil
}
