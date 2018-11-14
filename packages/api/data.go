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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

const binaryColumn = "data"

func compareHash(data []byte, urlHash string) bool {
	urlHash = strings.ToLower(urlHash)

	var hash []byte
	switch len(urlHash) {
	case 32:
		h := md5.Sum(data)
		hash = h[:]
	case 64:
		hash, _ = crypto.Hash(data)
	}

	return hex.EncodeToString(hash) == urlHash
}

func getDataHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	table, column := params["table"], params["column"]

	data, err := model.GetColumnByID(table, column, params["id"])
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting data from table")
		errorResponse(w, errNotFound)
		return
	}

	if !compareHash([]byte(data), params["hash"]) {
		errorResponse(w, errHashWrong)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(data))
	return
}

func getBinaryHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	bin := model.Binary{}
	found, err := bin.GetByID(converter.StrToInt64(params["id"]))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Errorf("getting binary by id")
		errorResponse(w, err)
		return
	}

	if !found {
		errorResponse(w, errNotFound)
		return
	}

	if !compareHash(bin.Data, params["hash"]) {
		errorResponse(w, errHashWrong)
		return
	}

	w.Header().Set("Content-Type", bin.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, bin.Name))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(bin.Data)
	return
}
