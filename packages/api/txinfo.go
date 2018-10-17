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
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/smart"

	log "github.com/sirupsen/logrus"
)

type txinfoResult struct {
	BlockID string        `json:"blockid"`
	Confirm int           `json:"confirm"`
	Data    *smart.TxInfo `json:"data,omitempty"`
}

type multiTxInfoResult struct {
	Results map[string]*txinfoResult `json:"results"`
}

func getTxInfo(txHash string, w http.ResponseWriter, cntInfo bool) (*txinfoResult, error) {
	var status txinfoResult
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, errorAPI(w, `E_HASHWRONG`, http.StatusBadRequest)
	}
	txStatus := &blockchain.TxStatus{}
	found, err := txStatus.Get(nil, hash)
	if err != nil {
		return nil, errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		return &status, nil
	}
	status.BlockID = converter.Int64ToStr(txStatus.BlockID)
	confirm := &blockchain.Confirmation{}
	found, err = confirm.Get(nil, hash)
	if err != nil {
		return nil, errorAPI(w, err, http.StatusInternalServerError)
	}
	if found {
		status.Confirm = int(confirm.Good)
	}
	if cntInfo {
		status.Data, err = smart.TransactionData(txStatus.BlockHash, hash)
		if err != nil {
			return nil, errorAPI(w, err, http.StatusInternalServerError)
		}
	}
	return &status, nil
}

func txinfo(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	status, err := getTxInfo(data.params[`hash`].(string), w, data.params[`contractinfo`].(int64) > 0)
	if err != nil {
		return err
	}
	data.result = &status
	return nil
}

func txinfoMulti(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	result := &multiTxInfoResult{}
	result.Results = map[string]*txinfoResult{}
	var request struct {
		Hashes []string `json:"hashes"`
	}
	if err := json.Unmarshal([]byte(data.params["data"].(string)), &request); err != nil {
		return errorAPI(w, `E_HASHWRONG`, http.StatusBadRequest)
	}
	for _, hash := range request.Hashes {
		status, err := getTxInfo(hash, w, data.params[`contractinfo`].(int64) > 0)
		if err != nil {
			return err
		}
		result.Results[hash] = status
	}
	data.result = result
	return nil
}
