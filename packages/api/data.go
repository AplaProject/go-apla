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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const binaryColumn = "data"

var errWrongHash = errors.New("Wrong hash")

func dataHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	// TODO убрать
	// if strings.Contains(table, model.BinaryTableSuffix) && column == binaryColumn {
	// 	binaryHandler(w, r)
	// 	return
	// }

	data, err := model.GetColumnByID(params["table"], params["column"], params["id"])
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting data from table")
		errorResponse(w, errNotFound, http.StatusNotFound)
		return
	}

	if fmt.Sprintf(`%x`, md5.Sum([]byte(data))) != strings.ToLower(params["hash"]) {
		logger.WithFields(log.Fields{"type": consts.InvalidObject, "error": errWrongHash}).Error("wrong hash")
		errorResponse(w, errNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(data))
	return
}

func binaryHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	bin := model.Binary{}
	bin.SetTablePrefix(params["prefix"])

	if found, err := bin.GetByID(converter.StrToInt64(params["id"])); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Errorf("getting binary by id")
		errorResponse(w, errServer, http.StatusInternalServerError)
		return
	} else if !found {
		errorResponse(w, errNotFound, http.StatusNotFound)
		return
	}

	if bin.Hash != strings.ToLower(params["hash"]) {
		logger.WithFields(log.Fields{"type": consts.InvalidObject, "error": errWrongHash}).Error("wrong hash")
		errorResponse(w, errNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", bin.MimeType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(bin.Data)
	return
}
