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

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/smart"

	log "github.com/sirupsen/logrus"
)

type getTestResult struct {
	Value string `json:"value"`
}

type signTestResult struct {
	Signature string `json:"signature"`
	Public    string `json:"pubkey"`
}

func getTest(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	data.result = &getTestResult{Value: smart.GetTestValue(data.params[`name`].(string))}
	return nil
}

type signTestForm struct {
	Form
	Private string `schema:"private"`
	Forsign string `schema:"forsign"`
}

func signTestHandler(w http.ResponseWriter, r *http.Request) {
	form := &signTestForm{}
	if ok := ParseForm(w, r, form); !ok {
		return
	}

	logger := getLogger(r)

	sign, err := crypto.Sign(form.Private, form.Forsign)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing data with private key")
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	priv, err := hex.DecodeString(form.Private)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": form.Private}).Error("decoding private from hex")
		errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	pub, err := crypto.PrivateToPublic(priv)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("converting private key to public")
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	jsonResponse(w, &signTestResult{
		Signature: hex.EncodeToString(sign),
		Public:    hex.EncodeToString(pub),
	})
}
