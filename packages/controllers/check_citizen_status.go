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
	//	"fmt"

	"github.com/DayLightProject/go-daylight/packages/script"
	"github.com/DayLightProject/go-daylight/packages/smart"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

const NCheckCitizen = `check_citizen_status`

type checkPage struct {
	Data     *CommonPage
	TxType   string
	TxTypeId int64
	Values   map[string]string
	Fields   []utils.FieldInfo
}

func init() {
	newPage(NCheckCitizen)
}

func (c *Controller) CheckCitizenStatus() (string, error) {
	var err error
	var lastId int64

	//test
	if len(c.r.FormValue(`last_id`)) > 0 {
		lastId = utils.StrToInt64(c.r.FormValue(`last_id`))
	}

	//	field, err := c.Single(`SELECT value FROM ` + c.StateIdStr + `_state_parameters where parameter='citizen_fields'`).String()
	vals, err := c.OneRow(`select * from "`+c.StateIdStr+`_citizenship_requests" where approved=0 AND id>? order by id`, lastId).String()
	if err != nil {
		return ``, err
	}
	fields := make([]utils.FieldInfo, 0)
	if len(vals) > 0 {
		//		vals[`publicKey`] = hex.EncodeToString([]byte(vals[`public`]))
		//		pubkey, _ := c.Single(`select public_key_0 from dlt_wallets where wallet_id=?`, vals[`dlt_wallet_id`]).Bytes()
		//		vals[`publicKey`] = hex.EncodeToString(pubkey)
		//var data map[string]interface{}
		/*		if err = json.Unmarshal([]byte(vals[`data`]), &data); err != nil {
				return ``, err
			}*/
		contract := smart.GetContract(`TXCitizenRequest`)
		for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
			if fitem.Type.String() == `string` {
				value := vals[`name`] //fitem.Name] //.(string)
				fields = append(fields, utils.FieldInfo{Name: fitem.Name, HtmlType: "textinput",
					TxType: fitem.Type.String(), Title: fitem.Name,
					Value: value})
			}
		}
		vals[`publicKey`] = hex.EncodeToString([]byte(vals[`public_key_0`])) //.(string)
	}
	txType := "TXNewCitizen"
	return proceedTemplate(c, NCheckCitizen, &checkPage{Data: c.Data, Values: vals,
		Fields: fields, TxType: txType, TxTypeId: utils.TypeInt(txType)})
}
