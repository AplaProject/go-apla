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

package controllers

import (
	"fmt"

	"github.com/DayLightProject/go-daylight/packages/utils"
)

const ACitizenFields = `ajax_citizen_fields`

type CitizenFieldsJson struct {
	Fields   string `json:"fields"`
	Price    int64  `json:"price"`
	Valid    bool   `json:"valid"`
	Approved int64  `json:"approved"`
	Error    string `json:"error"`
}

func init() {
	newPage(ACitizenFields, `json`)
}

func (c *Controller) AjaxCitizenFields() interface{} {
	var (
		result CitizenFieldsJson
		err    error
		amount int64
	)
	stateId := int64(1) // utils.StrToInt64(c.r.FormValue(`state_id`))
	//	_, err = c.GetStateName(stateId)
	//	if err == nil {
	if req, err := c.OneRow(`select id, approved from "`+utils.Int64ToStr(stateId)+`_citizenship_requests" where dlt_wallet_id=? order by id desc`,
		c.SessWalletId).Int64(); err == nil {
		if len(req) > 0 && req[`id`] > 0 {
			result.Approved = req[`approved`]
		} else {
			result.Fields, err = `[{"name":"name", "htmlType":"textinput", "txType":"string", "title":"First Name"},
{"name":"lastname", "htmlType":"textinput", "txType":"string", "title":"Last Name"},
{"name":"birthday", "htmlType":"calendar", "txType":"string", "title":"Birthday"},
{"name":"photo", "htmlType":"file", "txType":"binary", "title":"Photo"}
]`, nil
			//				c.Single(`SELECT value FROM ` + utils.Int64ToStr(stateId) + `_state_parameters where parameter='citizen_fields'`).String()
			if err == nil {
				result.Price, err = c.Single(`SELECT value FROM "` + utils.Int64ToStr(stateId) + `_state_parameters" where name='citizenship_price'`).Int64()
				if err == nil {
					amount, err = c.Single("select amount from dlt_wallets where wallet_id=?", c.SessWalletId).Int64()
					result.Valid = (err == nil && amount >= result.Price)
				}
			}
		}
	}
	fmt.Println(`Error`, err)
	//	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
