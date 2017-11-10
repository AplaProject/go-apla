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
	"encoding/hex"
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type txstatusResult struct {
	BlockID string `json:"blockid"`
	Message string `json:"errmsg"`
	Result  string `json:"result"`
}

func txstatus(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var status txstatusResult

	if _, err := hex.DecodeString(data.params[`hash`].(string)); err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err}).Error("decoding tx hash from hex")
		return errorAPI(w, `E_HASHWRONG`, http.StatusBadRequest)
	}
	ts := &model.TransactionStatus{}
	found, err := ts.Get([]byte(converter.HexToBin(data.params["hash"].(string))))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err}).Error("getting transaction status by hash")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": []byte(converter.HexToBin(data.params["hash"].(string)))}).Error("getting transaction status by hash")
		return errorAPI(w, `E_HASHNOTFOUND`, http.StatusBadRequest)
	}
	if ts.BlockID > 0 {
		status.BlockID = converter.Int64ToStr(ts.BlockID)
		status.Result = ts.Error
	} else {
		status.Message = ts.Error
	}
	data.result = &status
	return nil
}
