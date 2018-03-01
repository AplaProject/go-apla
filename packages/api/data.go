// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
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
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"

	hr "github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

const (
	base64columnType = "bytea"
	base64header     = "base64,"
)

var base64regexp = regexp.MustCompile(`(?is)^data:([a-z0-9-]+\/[a-z0-9-]+);base64,$`)

func dataHandler() hr.Handle {
	return hr.Handle(func(w http.ResponseWriter, r *http.Request, ps hr.Params) {
		tblname := ps.ByName("table")
		column := ps.ByName("column")

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

		columnType, err := model.GetColumnType(tblname, column)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting column type")
			errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
			return
		}

		if columnType != base64columnType {
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Write([]byte(data))
			return
		}

		offset := strings.Index(data, base64header)
		if offset == -1 {
			log.WithFields(log.Fields{"type": consts.InvalidObject, "error": fmt.Errorf("wrong data")}).Error("wrong data")
			errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
			return
		}
		ret := base64regexp.FindStringSubmatch(data[:offset+len(base64header)])
		if len(ret) != 2 {
			log.WithFields(log.Fields{"type": consts.InvalidObject, "error": fmt.Errorf("wrong data")}).Error("wrong data")
			errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
			return
		}

		datatype := ret[1]
		bin, err := base64.StdEncoding.DecodeString(data[offset+len(base64header):])
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("encoding base64")
			errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
		}
		w.Header().Set("Content-Type", datatype)
		w.Header().Set("Cache-Control", "public,max-age=604800,immutable")
		w.Write(bin)
		return
	})
}
