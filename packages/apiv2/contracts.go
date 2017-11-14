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

package apiv2

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"

	log "github.com/sirupsen/logrus"
)

type contractsResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

func getContracts(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var limit int

	table := getPrefix(data) + `_contracts`

	count, err := model.GetNextID(nil, table)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	if data.params[`limit`].(int64) > 0 {
		limit = int(data.params[`limit`].(int64))
	} else {
		limit = 25
	}
	list, err := model.GetAll(`select * from "`+table+`" order by id desc`+
		fmt.Sprintf(` offset %d `, data.params[`offset`].(int64)), limit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	for ind, val := range list {
		if val[`wallet_id`] == `NULL` {
			list[ind][`wallet_id`] = ``
			list[ind][`address`] = ``
		} else {
			list[ind][`address`] = converter.AddressToString(converter.StrToInt64(val[`wallet_id`]))
		}
		if val[`active`] == `NULL` {
			list[ind][`active`] = ``
		}
		list[ind][`name`] = strings.Join(smart.ContractsList(val[`value`]), `,`)
	}
	data.result = &listResult{
		Count: converter.Int64ToStr(count - 1), List: list,
	}
	return
}
