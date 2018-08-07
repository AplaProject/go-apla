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
	"io/ioutil"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type sendTxResult struct {
	Hash string `json:"hash"`
}

func sendTx(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	r.ParseMultipartForm(multipartBuf)
	file, _, err := r.FormFile("data")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("request.FormFile")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	defer file.Close()
	var txData []byte
	if txData, err = ioutil.ReadAll(file); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading multipart file")
		return err
	}
	if hash, err := model.SendTx(1, data.keyId, txData); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("sending tx")
		return errorAPI(w, err, http.StatusInternalServerError)
	} else {
		data.result = sendTxResult{Hash: string(converter.BinToHex(hash))}
	}
	return nil
}
