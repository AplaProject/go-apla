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
	"io/ioutil"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/block"
	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/queue"

	log "github.com/sirupsen/logrus"
)

type sendTxResult struct {
	Hashes map[string]string `json:"hashes"`
}

func getTxData(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry, key string) ([]byte, error) {
	file, _, err := r.FormFile(key)
	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("request.FormFile")
		return nil, errorAPI(w, err, http.StatusInternalServerError)
	}
	defer file.Close()

	var txData []byte
	if txData, err = ioutil.ReadAll(file); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading multipart file")
		return nil, err
	}

	return txData, nil
}

func sendTx(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	if block.IsKeyBanned(data.keyId) {
		return errorAPI(w, "E_BANNED", http.StatusBadRequest, block.BannedTill(data.keyId))
	}

	err := r.ParseMultipartForm(multipartBuf)
	if err != nil {
		return errorAPI(w, err, http.StatusBadRequest)
	}

	result := &sendTxResult{Hashes: make(map[string]string)}
	for key := range r.MultipartForm.File {
		txData, err := getTxData(w, r, data, logger, key)
		if err != nil {
			return err
		}

		hash, err := handlerTx(w, r, data, logger, txData)
		if err != nil {
			return err
		}
		result.Hashes[key] = hash
	}

	for key := range r.Form {
		txData, err := hex.DecodeString(r.FormValue(key))
		if err != nil {
			return err
		}

		hash, err := handlerTx(w, r, data, logger, txData)
		if err != nil {
			return err
		}
		result.Hashes[key] = hash
	}

	data.result = result

	return nil
}

type contractResult struct {
	Hash string `json:"hash"`
	// These fields are used for VDE
	Message *txstatusError `json:"errmsg,omitempty"`
	Result  string         `json:"result,omitempty"`
}

func handlerTx(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry, txData []byte) (string, error) {
	if int64(len(txData)) > syspar.GetMaxTxSize() {
		logger.WithFields(log.Fields{"type": consts.ParameterExceeded, "max_size": syspar.GetMaxTxSize(), "size": len(txData)}).Error("transaction size exceeds max size")
		block.BadTxForBan(data.keyId)
		return "", errorAPI(w, "E_LIMITTXSIZE", http.StatusBadRequest, len(txData))
	}

	rtx := &blockchain.Transaction{}
	if err := rtx.Unmarshal(txData); err != nil {
		return "", errorAPI(w, err, http.StatusInternalServerError)
	}

	if rtx.Header.KeyID != data.keyId {
		return "", errorAPI(w, "E_DIFKEY", http.StatusBadRequest)
	}
	if err := queue.ValidateTxQueue.Enqueue(rtx); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("sending tx")
		return "", errorAPI(w, err, http.StatusInternalServerError)
	}

	txHash, err := rtx.Hash()
	if err != nil {
		return "", errorAPI(w, err, http.StatusInternalServerError)
	}
	return string(converter.BinToHex(txHash)), nil
}
