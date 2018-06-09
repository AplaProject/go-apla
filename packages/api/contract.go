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

type contractRequest struct {
	Pubkey         string   `json:"pubkey"`
	TokenEcosystem string   `json:"token_ecosystem"`
	MaxSum         string   `json:"max_sum"`
	Payover        string   `json:"payover"`
	SignedBy       string   `json:"signed_by"`
	Signatures     []string `json:"signatures"`
	Time           string   `json:"time"`
}

type contractResult struct {
	Hashes []string `json:"hashes"`
}

func (c *contractHandlers) ContractMultiHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	client := getClient(r)
	logger := getLogger(r)

	requestID := params["request_id"]
	bufReq, ok := c.requests.GetRequest(requestID)
	if !ok {
		errorResponse(w, errRequestNotFound, http.StatusNotFound, requestID)
		return
	}

	req := contractRequest{}
	if err := json.Unmarshal([]byte(r.FormValue("data")), &req); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	var signedBy int64
	signID := client.KeyID
	if req.SignedBy != "" {
		signedBy = converter.StrToInt64(req.SignedBy)
		signID = signedBy
	}

	var publicKey []byte
	var err error
	pubkey := []byte{}
	if req.Pubkey != "" {
		pubkey, err = hex.DecodeString(req.Pubkey)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("converting signature from hex")
			errorResponse(w, err, http.StatusBadRequest)
			return
		}
	}
	publicKey, ok = getPublicKey(w, r, signID, client.EcosystemID, pubkey)
	if !ok {
		return
	}

	signatures := req.Signatures
	if len(signatures) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("signatures is empty")
		errorResponse(w, errEmptySign, http.StatusBadRequest)
		return
	}

	smartTx := newTxContract()
	smartTx.RequestID = bufReq.ID
	smartTx.TokenEcosystem = converter.StrToInt64(req.TokenEcosystem)
	smartTx.MaxSum = req.MaxSum
	smartTx.PayOver = req.Payover
	smartTx.SignedBy = signedBy

	hashes := []string{}
	for i, contReq := range bufReq.Contracts {
		contract := getContract(r, contReq.Contract())
		if contract == nil {
			errorResponse(w, errContract, http.StatusBadRequest, contReq.Contract())
			return
		}

		signatureBytes, err := hex.DecodeString(signatures[i])
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("converting signature from hex")
			errorResponse(w, err, http.StatusInternalServerError)
			return
		}

		smartTx.Header = newTxHeader()
		smartTx.Header.Time = converter.StrToInt64(req.Time)
		smartTx.Header.EcosystemID = client.EcosystemID
		smartTx.Header.KeyID = client.KeyID
		smartTx.Header.RoleID = client.RoleID
		smartTx.Header.NetworkID = consts.NETWORK_ID
		smartTx.Header.PublicKey = publicKey
		smartTx.Header.BinSignatures = converter.EncodeLengthPlusData(signatureBytes)

		hash, err := contract.CreateTxFromRequest(contReq, smartTx)
		if err != nil {
			errorResponse(w, err, http.StatusBadRequest)
			return
		}

		hashes = append(hashes, hash)
	}

	jsonResponse(w, &contractResult{
		Hashes: hashes,
	})
}

func getPublicKey(w http.ResponseWriter, r *http.Request, signID int64, ecosystemID int64, pubkey []byte) ([]byte, bool) {
	logger := getLogger(r)

	var publicKey []byte
	key := &model.Key{}
	key.SetTablePrefix(ecosystemID)
	_, err := key.Get(signID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting public key from keys")
		errorResponse(w, err, http.StatusInternalServerError)
		return nil, false
	}
	if key.Deleted == 1 {
		errorResponse(w, errDeletedKey, http.StatusForbidden)
		return nil, false
	}
	if len(key.PublicKey) == 0 {
		if len(pubkey) > 0 {
			publicKey = pubkey
			lenpub := len(publicKey)
			if lenpub > 64 {
				publicKey = publicKey[lenpub-64:]
			}
		}
		if len(publicKey) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty")
			errorResponse(w, errEmptyPublic, http.StatusBadRequest)
			return nil, false
		}
	} else {
		logger.Warning("public key for wallet not found")
		publicKey = []byte("null")
	}
	return publicKey, true
}
