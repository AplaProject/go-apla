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
	"encoding/hex"
	"encoding/json"

	"github.com/DayLightProject/go-daylight/packages/utils"
)

//	"fmt"

const NCheckCitizen = `check_citizen_status`

type checkPage struct {
	Data     *CommonPage
	TxType   string
	TxTypeId int64
	Values   map[string]string
	Fields   []FieldInfo
}

func init() {
	newPage(NCheckCitizen)
}

func (c *Controller) CheckCitizenStatus() (string, error) {
	var fields []FieldInfo
	var err error

	if len(c.r.FormValue(`accept`)) > 0 {
		requestId := utils.StrToInt64(c.r.FormValue(`request_id`))
		approved := -1
		if c.r.FormValue(`accept`) == `true` {
			approved = 1
		}
		if err := c.ExecSql(`update `+c.StateIdStr+`_citizenship_requests set approved=? where id=?`,
			approved, requestId); err != nil {
			return ``, err
		}
	}
	//	field, err := c.Single(`SELECT value FROM ` + c.StateIdStr + `_state_parameters where parameter='citizen_fields'`).String()
	field, err := `[{"name":"name", "htmlType":"textinput", "txType":"string", "title":"First Name"},
{"name":"lastname", "htmlType":"textinput", "txType":"string", "title":"Last Name"},
{"name":"birthday", "htmlType":"calendar", "txType":"string", "title":"Birthday"},
{"name":"photo", "htmlType":"file", "txType":"binary", "title":"Photo"}
]`, nil

	if err != nil {
		return ``, err
	}
	if err = json.Unmarshal([]byte(field), &fields); err != nil {
		return ``, err
	}
	vals, err := c.OneRow(`select * from ` + c.StateIdStr + `_citizenship_requests where approved=0 order by id`).String()
	if err != nil {
		return ``, err
	}
	if len(vals) > 0 {
		//		vals[`publicKey`] = hex.EncodeToString([]byte(vals[`public`]))
		pubkey, _ := c.Single(`select public_key_0 from dlt_wallets where wallet_id=?`, vals[`dlt_wallet_id`]).Bytes()
		vals[`publicKey`] = hex.EncodeToString(pubkey)
	}
	txType := "NewCitizen"
	return proceedTemplate(c, NCheckCitizen, &checkPage{Data: c.Data, Values: vals,
		Fields: fields, TxType: txType, TxTypeId: utils.TypeInt(txType)})
}
