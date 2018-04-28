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
	"encoding/json"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const keyHash = "hash"

type txstatusError struct {
	Type  string `json:"type,omitempty"`
	Error string `json:"error,omitempty"`
}

type txstatusResult struct {
	BlockID string         `json:"blockid"`
	Message *txstatusError `json:"errmsg,omitempty"`
	Result  string         `json:"result"`
}

func txstatusHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	status := &txstatusResult{}

	if _, err := hex.DecodeString(params[keyHash]); err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding tx hash from hex")
		errorResponse(w, errHashWrong, http.StatusBadRequest)
		return
	}
	ts := &model.TransactionStatus{}
	hash := []byte(converter.HexToBin(params[keyHash]))
	if found, err := ts.Get(hash); err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("getting transaction status by hash")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	} else if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": hash}).Error("getting transaction status by hash")
		errorResponse(w, `E_HASHNOTFOUND`, http.StatusBadRequest)
		return
	}
	if ts.BlockID > 0 {
		status.BlockID = converter.Int64ToStr(ts.BlockID)
		status.Result = ts.Error
	} else if len(ts.Error) > 0 {
		if err := json.Unmarshal([]byte(ts.Error), &status.Message); err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "text": ts.Error, "error": err}).Warn("unmarshalling txstatus error")
			status.Message = &txstatusError{
				Type:  "txError",
				Error: ts.Error,
			}
		}
	}

	jsonResponse(w, status)
}
