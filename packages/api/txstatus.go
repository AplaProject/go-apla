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

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type txstatusError struct {
	Type  string `json:"type,omitempty"`
	Error string `json:"error,omitempty"`
}

type txstatusResult struct {
	BlockID string         `json:"blockid"`
	Message *txstatusError `json:"errmsg,omitempty"`
	Result  string         `json:"result"`
}

func getTxStatus(hash string, w http.ResponseWriter, logger *log.Entry) (*txstatusResult, error) {
	var status txstatusResult
	if _, err := hex.DecodeString(hash); err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding tx hash from hex")
		return nil, errorAPI(w, `E_HASHWRONG`, http.StatusBadRequest)
	}
	ts := &model.TransactionStatus{}
	found, err := ts.Get([]byte(converter.HexToBin(hash)))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("getting transaction status by hash")
		return nil, errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": []byte(converter.HexToBin(hash))}).Error("getting transaction status by hash")
		return nil, errorAPI(w, `E_HASHNOTFOUND`, http.StatusBadRequest)
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
	return &status, nil
}

type multiTxStatusResult struct {
	Results map[string]*txstatusResult `json:"results"`
}

func txstatus(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	status, err := getTxStatus(data.params[`hash`].(string), w, logger)
	if err != nil {
		return err
	}
	data.result = &status
	return nil
}

func txstatusMulti(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	result := &multiTxStatusResult{}
	result.Results = map[string]*txstatusResult{}
	var request struct {
		Hashes []string `json:"hashes"`
	}
	if err := json.Unmarshal([]byte(data.params["data"].(string)), &request); err != nil {
		return errorAPI(w, `E_HASHWRONG`, http.StatusBadRequest)
	}
	for _, hash := range request.Hashes {
		status, err := getTxStatus(hash, w, logger)
		if err != nil {
			return err
		}
		result.Results[hash] = status
	}
	data.result = result
	return nil
}
