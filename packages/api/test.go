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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/smart"

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

func signTest(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {

	sign, err := crypto.Sign(data.params[`private`].(string), data.params[`forsign`].(string))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing data with private key")
		return errorAPI(w, err, http.StatusBadRequest)
	}
	private, err := hex.DecodeString(data.params[`private`].(string))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": data.params["private"].(string)}).Error("decoding private from hex")
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	pub, err := crypto.PrivateToPublic(private)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("converting private key to public")
		return errorAPI(w, err, http.StatusBadRequest)
	}
	data.result = &signTestResult{Signature: hex.EncodeToString(sign), Public: hex.EncodeToString(pub)}
	return nil
}
