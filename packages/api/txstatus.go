//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
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

func txstatus(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var status txstatusResult

	if _, err := hex.DecodeString(data.params[`hash`].(string)); err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding tx hash from hex")
		return errorAPI(w, `E_HASHWRONG`, http.StatusBadRequest)
	}
	ts := &model.TransactionStatus{}
	found, err := ts.Get([]byte(converter.HexToBin(data.params["hash"].(string))))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("getting transaction status by hash")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": []byte(converter.HexToBin(data.params["hash"].(string)))}).Error("getting transaction status by hash")
		return errorAPI(w, `E_HASHNOTFOUND`, http.StatusBadRequest)
	}
	if ts.BlockID > 0 {
		status.BlockID = converter.Int64ToStr(ts.BlockID)
		status.Result = ts.Error
	} else if len(ts.Error) > 0 {
		if err := json.Unmarshal([]byte(ts.Error), &status.Message); err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "text": ts.Error,
				"error": err}).Error("unmarshalling txstatus error")
			return errorAPI(w, err, http.StatusInternalServerError)
		}
	}
	data.result = &status
	return nil
}
