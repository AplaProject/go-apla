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
	"time"

	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/script"
	"github.com/GenesisCommunity/go-genesis/packages/utils/tx"

	log "github.com/sirupsen/logrus"
)

type prepareResult struct {
	ForSign string            `json:"forsign"`
	Signs   []TxSignJSON      `json:"signs"`
	Values  map[string]string `json:"values"`
	Time    string            `json:"time"`
}

func prepareContract(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var (
		result  prepareResult
		timeNow int64
		smartTx tx.SmartContract
	)

	timeNow = time.Now().Unix()
	result.Time = converter.Int64ToStr(timeNow)
	result.Values = make(map[string]string)
	contract, parerr, err := validateSmartContract(r, data, &result)
	if err != nil {
		if strings.HasPrefix(err.Error(), `E_`) {
			return errorAPI(w, err.Error(), http.StatusBadRequest, parerr)
		}
		return errorAPI(w, err, http.StatusBadRequest)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	smartTx.TokenEcosystem = data.params[`token_ecosystem`].(int64)
	smartTx.MaxSum = data.params[`max_sum`].(string)
	smartTx.PayOver = data.params[`payover`].(string)
	if data.params[`signed_by`] != nil {
		smartTx.SignedBy = data.params[`signed_by`].(int64)
	}
	smartTx.Header = tx.Header{Type: int(info.ID), Time: timeNow, EcosystemID: data.ecosystemId, KeyID: data.keyId}
	forsign := smartTx.ForSign()
	if info.Tx != nil {
		for _, fitem := range *info.Tx {
			if strings.Contains(fitem.Tags, `image`) || strings.Contains(fitem.Tags, `signature`) {
				continue
			}
			var val string
			if fitem.Type.String() == `[]interface {}` {
				for key, values := range r.Form {
					if key == fitem.Name+`[]` {
						var list []string
						for _, value := range values {
							list = append(list, value)
						}
						val = strings.Join(list, `,`)
					}
				}
			} else {
				val = strings.TrimSpace(r.FormValue(fitem.Name))
				if strings.Contains(fitem.Tags, `address`) {
					val = converter.Int64ToStr(converter.StringToAddress(val))
				} else if fitem.Type.String() == script.Decimal {
					val = strings.TrimLeft(val, `0`)
				} else if fitem.Type.String() == `int64` && len(val) == 0 {
					val = `0`
				}
			}
			forsign += fmt.Sprintf(",%v", val)
		}
	}
	result.ForSign = forsign
	data.result = result
	return nil
}
