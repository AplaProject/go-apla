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
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"

	hr "github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

func dataHandler() hr.Handle {
	return hr.Handle(func(w http.ResponseWriter, r *http.Request, ps hr.Params) {
		var data string
		err := model.DBConn.Table(ps.ByName(`table`)).Select(ps.ByName(`column`)).Where(`id=?`,
			ps.ByName(`id`)).Row().Scan(&data)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting data from table")
			errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
			return
		}
		mark := `base64,`
		offset := strings.Index(data, mark)
		if offset == -1 || fmt.Sprintf(`%x`, md5.Sum([]byte(data))) != strings.ToLower(ps.ByName(`hash`)) {
			return
		}
		re := regexp.MustCompile(`(?is)^data:([a-z0-9-]+\/[a-z0-9-]+);base64,$`)
		ret := re.FindStringSubmatch(data[:offset+len(mark)])
		if len(ret) != 2 {
			return
		}
		datatype := ret[1]
		bin, err := base64.StdEncoding.DecodeString(data[offset+len(mark):])
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
