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
	"encoding/hex"
	"net/http"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
)

type txstatusResult struct {
	BlockID string `json:"blockid"`
	Message string `json:"errmsg"`
}

func txstatus(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var status txstatusResult

	if _, err := hex.DecodeString(data.params[`hash`].(string)); err != nil {
		return errorAPI(w, `hash is incorrect`, http.StatusBadRequest)
	}
	tx, err := sql.DB.OneRow(`SELECT block_id, error FROM transactions_status WHERE hash = [hex]`,
		data.params[`hash`].(string)).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	if len(tx) == 0 {
		return errorAPI(w, `hash has not been found`, http.StatusBadRequest)
	}
	if converter.StrToInt64(tx[`block_id`]) > 0 {
		status.BlockID = tx[`block_id`]
	}
	status.Message = tx[`error`]
	data.result = &status
	return nil
}
