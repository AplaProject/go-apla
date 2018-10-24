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

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"

	log "github.com/sirupsen/logrus"
)

type txstatusError struct {
	Type  string `json:"type,omitempty"`
	Error string `json:"error,omitempty"`
	Id    string `json:"id,omitempty"`
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
	tx := &blockchain.TxStatus{}
	found, err := tx.Get(nil, converter.HexToBin(hash))
	if err != nil {
		return nil, errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": []byte(converter.HexToBin(hash))}).Error("getting transaction status by hash")
		return nil, errorAPI(w, `E_HASHNOTFOUND`, http.StatusBadRequest)
	}
	if err != nil {
		return nil, errorAPI(w, err, http.StatusInternalServerError)
	}
	if tx.BlockID > 0 {
		status.BlockID = converter.Int64ToStr(tx.BlockID)
		status.Result = tx.Error
	} else if len(tx.Error) > 0 {
		if err := json.Unmarshal([]byte(tx.Error), &status.Message); err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "text": tx.Error, "error": err}).Warn("unmarshalling txstatus error")
			status.Message = &txstatusError{
				Type:  "txError",
				Error: tx.Error,
			}
		}
	}
	return &status, nil
}

type multiTxStatusResult struct {
	Results map[string]*txstatusResult `json:"results"`
}

type txstatusRequest struct {
	Hashes []string `json:"hashes"`
}

func txstatus(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	result := &multiTxStatusResult{}
	result.Results = map[string]*txstatusResult{}
	var request txstatusRequest
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
