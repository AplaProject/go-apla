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
	"fmt"
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"

	log "github.com/sirupsen/logrus"
)

const (
	defaultPaginatorLimit = 25
)

type contractsResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

type paginatorForm struct {
	form
	Limit  int   `schema:"limit"`
	Offset int64 `schema:"offset"`
}

func (f *paginatorForm) Validate(r *http.Request) error {
	if f.Limit <= 0 {
		f.Limit = defaultPaginatorLimit
	}
	return nil
}

func contractsHandler(w http.ResponseWriter, r *http.Request) {
	form := &paginatorForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err)
		return
	}

	client := getClient(r)
	logger := getLogger(r)

	table := client.Prefix() + "_contracts"
	count, err := model.GetRecordsCountTx(nil, table)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id")
		errorResponse(w, err)
		return
	}

	// TODO: перенести запрос в модели
	list, err := model.GetAll(`select * from "`+table+`" order by id desc`+
		fmt.Sprintf(` offset %d `, form.Offset), form.Limit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all")
		errorResponse(w, err)
		return
	}
	for ind, val := range list {
		list[ind]["address"] = converter.AddressToString(converter.StrToInt64(val["wallet_id"]))
		cntlist, err := script.ContractsList(val["value"])
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ContractError, "error": err}).Error("getting contract list")
			errorResponse(w, err)
			return
		}
		list[ind]["name"] = strings.Join(cntlist, `,`)
	}

	jsonResponse(w, &listResult{
		Count: converter.Int64ToStr(count),
		List:  list,
	})
	return
}
