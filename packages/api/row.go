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
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type rowResult struct {
	Value map[string]string `json:"value"`
}

func row(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	cols := `*`
	if len(data.params[`columns`].(string)) > 0 {
		cols = converter.EscapeName(data.params[`columns`].(string))
	}
	table := converter.EscapeName(getPrefix(data) + `_` + data.params[`name`].(string))
	row, err := model.GetOneRow(`SELECT `+cols+` FROM `+table+` WHERE id = ?`, data.params[`id`].(string)).String()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": data.params["name"].(string), "id": data.params["id"].(string)}).Error("getting one row")
		return errorAPI(w, `E_QUERY`, http.StatusInternalServerError)
	}

	data.result = &rowResult{Value: row}
	return
}
