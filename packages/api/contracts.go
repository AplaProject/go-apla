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

type contractsResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

func getContracts(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var limit int

	table := `1_contracts`

	count, err := model.GetRecordsCountTx(nil, table, ``)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting table records count")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	if data.params[`limit`].(int64) > 0 {
		limit = int(data.params[`limit`].(int64))
	} else {
		limit = 25
	}
	list, err := model.GetAll(fmt.Sprintf(`select * from "%s" order by id desc offset %d `, table, data.params[`offset`].(int64)), limit)
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
		cntlist, err := script.ContractsList(val[`value`])
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ContractError, "error": err}).Error("getting contract list")
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		list[ind][`name`] = strings.Join(cntlist, `,`)
	}
	data.result = &listResult{
		Count: converter.Int64ToStr(count), List: list,
	}
	return
}
