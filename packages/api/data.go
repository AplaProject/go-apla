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
	"fmt"
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	hr "github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

const binaryColumn = "data"

func dataHandler() hr.Handle {
	return hr.Handle(func(w http.ResponseWriter, r *http.Request, ps hr.Params) {
		tblname := ps.ByName("table")
		column := ps.ByName("column")

		if strings.Contains(tblname, model.BinaryTableSuffix) && column == binaryColumn {
			binary(w, r, ps)
			return
		}

		data, err := model.GetColumnByID(tblname, column, ps.ByName(`id`))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting data from table")
			errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
			return
		}

		if fmt.Sprintf(`%x`, md5.Sum([]byte(data))) != strings.ToLower(ps.ByName(`hash`)) {
			log.WithFields(log.Fields{"type": consts.InvalidObject, "error": fmt.Errorf("wrong hash")}).Error("wrong hash")
			errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(data))
		return
	})
}

func binary(w http.ResponseWriter, r *http.Request, ps hr.Params) {
	bin := model.Binary{}
	bin.SetTableName(ps.ByName("table"))

	found, err := bin.GetByID(converter.StrToInt64(ps.ByName("id")))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Errorf("getting binary by id")
		errorAPI(w, "E_SERVER", http.StatusInternalServerError)
		return
	}

	if !found {
		errorAPI(w, "E_SERVER", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", bin.MimeType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(bin.Data)
	return
}
