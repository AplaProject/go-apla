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

type tableInfo struct {
	Name  string `json:"name"`
	Count string `json:"count"`
}

type tablesResult struct {
	Count int64       `json:"count"`
	List  []tableInfo `json:"list"`
}

func getTablesHandler(w http.ResponseWriter, r *http.Request) {
	form := &paginatorForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadGateway)
		return
	}

	client := getClient(r)
	logger := getLogger(r)
	prefix := client.Prefix()

	table := &model.Table{}
	table.SetTablePrefix(prefix)

	count, err := table.Count()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting records count from tables")
		errorResponse(w, err)
		return
	}

	rows, err := model.GetDB(nil).Table(table.TableName()).Where("ecosystem = ?", client.EcosystemID).Offset(form.Offset).Limit(form.Limit).Rows()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting rows from table")
		errorResponse(w, err)
		return
	}

	list, err := model.GetResult(rows)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting names from tables")
		errorResponse(w, err)
		return
	}

	result := &tablesResult{
		Count: count,
		List:  make([]tableInfo, len(list)),
	}
	for i, item := range list {
		err = model.GetTableQuery(item["name"], client.EcosystemID).Count(&count).Error
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting count from table")
			errorResponse(w, err)
			return
		}

		result.List[i].Name = item["name"]
		result.List[i].Count = converter.Int64ToStr(count)
	}

	jsonResponse(w, result)
}
