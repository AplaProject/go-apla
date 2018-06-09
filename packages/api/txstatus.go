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

func getTxStatus(w http.ResponseWriter, r *http.Request, hash string) (*txstatusResult, bool) {
	var status txstatusResult
	logger := getLogger(r)
	if _, err := hex.DecodeString(hash); err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding tx hash from hex")
		errorResponse(w, errWrongHash, http.StatusBadRequest)
		return nil, false
	}
	ts := &model.TransactionStatus{}
	found, err := ts.Get([]byte(converter.HexToBin(hash)))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("getting transaction status by hash")
		errorResponse(w, err, http.StatusInternalServerError)
		return nil, false
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": []byte(converter.HexToBin(hash))}).Error("getting transaction status by hash")
		errorResponse(w, errHashNotFound, http.StatusBadRequest)
		return nil, false
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
	return &status, true
}

type multiTxStatusResult struct {
	Results map[string]*txstatusResult `json:"results"`
}

type multiTxStatusForm struct {
	Form
	Data string `schema:"data"`
}

func (f *multiTxStatusForm) Hashes() ([]string, error) {
	var result struct {
		Hashes []string `json:"hashes"`
	}

	if err := json.Unmarshal([]byte(f.Data), &result); err != nil {
		return nil, err
	}

	return result.Hashes, nil
}

// func txstatusHandler(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)

// 	status, ok := getTxStatus(w, r, params[keyHash])
// 	if !ok {
// 		return
// 	}
// 	jsonResponse(w, status)
// }

func txstatusMultiHandler(w http.ResponseWriter, r *http.Request) {
	form := &multiTxStatusForm{}
	if ok := ParseForm(w, r, form); !ok {
		return
	}

	result := &multiTxStatusResult{}
	result.Results = map[string]*txstatusResult{}

	hashes, err := form.Hashes()
	if err != nil {
		errorResponse(w, errHashWrong, http.StatusBadRequest)
		return
	}

	for _, hash := range hashes {
		status, ok := getTxStatus(w, r, hash)
		if !ok {
			return
		}
		result.Results[hash] = status
	}

	jsonResponse(w, result)
}
