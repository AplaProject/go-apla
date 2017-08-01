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

import "github.com/EGaaS/go-egaas-mvp/packages/model"
import "strconv"

const aCitizenFields = `ajax_citizen_fields`

// CitizenFieldsJSON is a structure for the answer of ajax_citizen_fields ajax request
type CitizenFieldsJSON struct {
	Fields   string `json:"fields"`
	Price    int64  `json:"price"`
	Valid    bool   `json:"valid"`
	Approved int64  `json:"approved"`
	Error    string `json:"error"`
}

func init() {
	newPage(aCitizenFields, `json`)
}

// AjaxCitizenFields is a controller of ajax_citizen_fields request
func (c *Controller) AjaxCitizenFields() interface{} {
	var (
		result CitizenFieldsJSON
		err    error
		amount int64
	)
	stateID := int64(1) // utils.StrToInt64(c.r.FormValue(`state_id`))
	//	_, err = c.GetStateName(stateId)
	//	if err == nil {
	request := &model.CitizenshipRequests{}
	request.SetTableName(stateID)
	err = request.GetByWallet(c.SessWalletID)
	if err == nil {
		if request.ID > 0 {
			result.Approved = request.Approved
		} else {
			result.Fields, err = `[{"name":"name", "htmlType":"textinput", "txType":"string", "title":"First Name"},
{"name":"lastname", "htmlType":"textinput", "txType":"string", "title":"Last Name"},
{"name":"birthday", "htmlType":"calendar", "txType":"string", "title":"Birthday"},
{"name":"photo", "htmlType":"file", "txType":"binary", "title":"Photo"}
]`, nil
			stateParameters := &model.StateParameters{}
			stateParameters.SetTableName(stateID)
			err = stateParameters.GetByName("citizenship_price")
			if err == nil {
				price, err := strconv.ParseInt(stateParameters.Value, 10, 64)
				if err != nil {
					result.Error = err.Error()
				} else {
					result.Price = price
					wallet := &model.DltWallet{}
					err = wallet.GetWallet(c.SessWalletID)
					result.Valid = (err == nil && amount >= result.Price)
				}
			}
		}
	} else {
		result.Error = err.Error()
	}
	return result
}
