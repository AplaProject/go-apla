// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package api

import (
	"encoding/hex"
	"net/http"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/crypto"
	"github.com/GenesisCommunity/go-genesis/packages/smart"

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
