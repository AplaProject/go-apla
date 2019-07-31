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
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"

	"github.com/gorilla/mux"
)

type txinfoResult struct {
	BlockID string        `json:"blockid"`
	Confirm int           `json:"confirm"`
	Data    *smart.TxInfo `json:"data,omitempty"`
}

type txInfoForm struct {
	nopeValidator
	ContractInfo bool   `schema:"contractinfo"`
	Data         string `schema:"data"`
}

type multiTxInfoResult struct {
	Results map[string]*txinfoResult `json:"results"`
}

func getTxInfo(r *http.Request, txHash string, cntInfo bool) (*txinfoResult, error) {
	var status txinfoResult
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, errHashWrong
	}
	ltx := &model.LogTransaction{Hash: hash}
	found, err := ltx.GetByHash(hash)
	if err != nil {
		return nil, err
	}
	if !found {
		return &status, nil
	}
	status.BlockID = converter.Int64ToStr(ltx.Block)
	var confirm model.Confirmation
	found, err = confirm.GetConfirmation(ltx.Block)
	if err != nil {
		return nil, err
	}
	if found {
		status.Confirm = int(confirm.Good)
	}
	if cntInfo {
		status.Data, err = smart.TransactionData(ltx.Block, hash)
		if err != nil {
			return nil, err
		}
	}
	return &status, nil
}

func getTxInfoHandler(w http.ResponseWriter, r *http.Request) {
	form := &txInfoForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	status, err := getTxInfo(r, params["hash"], form.ContractInfo)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, status)
}

func getTxInfoMultiHandler(w http.ResponseWriter, r *http.Request) {
	form := &txInfoForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	result := &multiTxInfoResult{}
	result.Results = map[string]*txinfoResult{}
	var request struct {
		Hashes []string `json:"hashes"`
	}
	if err := json.Unmarshal([]byte(form.Data), &request); err != nil {
		errorResponse(w, errHashWrong)
		return
	}
	for _, hash := range request.Hashes {
		status, err := getTxInfo(r, hash, form.ContractInfo)
		if err != nil {
			errorResponse(w, err)
			return
		}
		result.Results[hash] = status
	}

	jsonResponse(w, result)
}
