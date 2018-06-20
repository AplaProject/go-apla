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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
)

const multipartFormMaxMemory = 32 << 20 // 32 MB

type prepareResult struct {
	ID       string   `json:"request_id"`
	ForSigns []string `json:"forsign"`
	Time     string   `json:"time"`
}

type prepareRequest struct {
	TokenEcosystem string `json:"token_ecosystem"`
	MaxSum         string `json:"max_sum"`
	Payover        string `json:"payover"`
	SignedBy       string `json:"signed_by"`

	Contracts []prepareRequestItem `json:"contracts"`
}

type prepareRequestItem struct {
	Contract string            `json:"contract"`
	Params   map[string]string `json:"params"`
}

func (p prepareRequestItem) Get(key string) string {
	return p.Params[key]
}

type contractHandlers struct {
	requests *tx.RequestBuffer
}

func (h *contractHandlers) PrepareHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(multipartFormMaxMemory)

	req := prepareRequest{}
	if err := json.Unmarshal([]byte(r.FormValue("data")), &req); err != nil {
		fmt.Println(r.FormValue("data"))
		errorResponse(w, newError(err, http.StatusBadRequest))
		return
	}

	bufReq := h.requests.NewRequest()
	now := bufReq.Time.Unix()

	smartTx := newTxContract()
	smartTx.RequestID = bufReq.ID
	smartTx.TokenEcosystem = converter.StrToInt64(req.TokenEcosystem)
	smartTx.MaxSum = req.MaxSum
	smartTx.PayOver = req.Payover

	if req.SignedBy != "" {
		smartTx.SignedBy = converter.StrToInt64(req.SignedBy)
	}

	client := getClient(r)

	forSigns := []string{}
	for _, rc := range req.Contracts {
		contract := getContract(r, rc.Contract)
		if contract == nil {
			errorResponse(w, errContract.Errorf(rc.Contract))
			return
		}

		if err := contract.ValidateParams(rc); err != nil {
			errorResponse(w, newError(err, http.StatusBadRequest))
			return
		}

		smartTx.Header = newTxHeader()
		smartTx.Header.Time = now
		smartTx.Header.EcosystemID = client.EcosystemID
		smartTx.Header.KeyID = client.KeyID
		smartTx.Header.RoleID = client.RoleID
		smartTx.Header.NetworkID = consts.NETWORK_ID

		forSign, err := contract.ForSign(bufReq, smartTx, rc)
		if err != nil {
			errorResponse(w, newError(err, http.StatusBadRequest))
			return
		}

		forSigns = append(forSigns, forSign)
	}
	h.requests.AddRequest(bufReq)

	jsonResponse(w, &prepareResult{
		ID:       bufReq.ID,
		ForSigns: forSigns,
		Time:     converter.Int64ToStr(now),
	})
}
