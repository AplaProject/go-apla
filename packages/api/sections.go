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
	"encoding/json"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/language"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

const defaultSectionsLimit = 100

type sectionsForm struct {
	paginatorForm
	Lang string `schema:"lang"`
}

func (f *sectionsForm) Validate(r *http.Request) error {
	if err := f.paginatorForm.Validate(r); err != nil {
		return err
	}

	if len(f.Lang) == 0 {
		f.Lang = r.Header.Get("Accept-Language")
	}

	return nil
}

func getSectionsHandler(w http.ResponseWriter, r *http.Request) {
	form := &sectionsForm{}
	form.defaultLimit = defaultSectionsLimit
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	client := getClient(r)
	logger := getLogger(r)

	table := "1_section"
	q := model.GetDB(nil).Table(table).Where("ecosystem = ? AND status > 0", client.EcosystemID).Order("id ASC")

	result := new(listResult)
	err := q.Count(&result.Count).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting table records count")
		errorResponse(w, errTableNotFound.Errorf(table))
		return
	}

	rows, err := q.Offset(form.Offset).Limit(form.Limit).Rows()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting rows from table")
		errorResponse(w, err)
		return
	}

	result.List, err = model.GetResult(rows)
	if err != nil {
		errorResponse(w, err)
		return
	}

	var sections []map[string]string
	for _, item := range result.List {
		var roles []int64
		if err := json.Unmarshal([]byte(item["roles_access"]), &roles); err != nil {
			errorResponse(w, err)
			return
		}
		if len(roles) > 0 {
			var added bool
			for _, v := range roles {
				if v == client.RoleID {
					added = true
					break
				}
			}
			if !added {
				continue
			}
		}

		item["title"] = language.LangMacro(item["title"], int(client.EcosystemID), form.Lang)
		sections = append(sections, item)
	}
	result.List = sections

	jsonResponse(w, result)
}
