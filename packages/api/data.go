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
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/model"

	hr "github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

const binaryColumn = "data"

var errWrongHash = errors.New("Wrong hash")

func dataHandler() hr.Handle {
	return hr.Handle(func(w http.ResponseWriter, r *http.Request, ps hr.Params) {
		tblname := ps.ByName("table")
		column := ps.ByName("column")

		if strings.Contains(tblname, model.BinaryTableSuffix) && column == binaryColumn {
			binary(w, r, ps)
			return
		}

		id := ps.ByName(`id`)
		data, err := model.GetColumnByID(tblname, column, id)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting data from table")
			errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
			return
		}

		if fmt.Sprintf(`%x`, md5.Sum([]byte(data))) != strings.ToLower(ps.ByName(`hash`)) {
			log.WithFields(log.Fields{"type": consts.InvalidObject, "error": errWrongHash}).Error("wrong hash")
			errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment")
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

	if bin.Hash != strings.ToLower(ps.ByName("hash")) {
		log.WithFields(log.Fields{"type": consts.InvalidObject, "error": errWrongHash}).Error("wrong hash")
		errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", bin.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, bin.Name))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(bin.Data)
	return
}
