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

	//	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/template"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const nCheckCitizen = `check_citizen_status`

type checkPage struct {
	Data     *CommonPage
	TxType   string
	TxTypeID int64
	Values   map[string]string
	Fields   []template.FieldInfo
}

func init() {
	newPage(nCheckCitizen)
}

// CheckCitizenStatus is controller for changing citizen status
func (c *Controller) CheckCitizenStatus() (string, error) {
	var err error
	var lastID int64

	//test
	if len(c.r.FormValue(`last_id`)) > 0 {
		lastID = converter.StrToInt64(c.r.FormValue(`last_id`))
	}
	request := &model.CitizenshipRequests{}
	request.SetTableName(c.StateID)
	err = request.GetUnapproved(lastID)
	if err != nil {
		return ``, err
	}
	fields := make([]template.FieldInfo, 0)
	contract := smart.GetContract(`TXCitizenRequest`, uint32(converter.StrToUint64(c.StateIDStr)))
	for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
		if fitem.Type.String() == `string` {
			value := request.Name
			fields = append(fields, template.FieldInfo{Name: fitem.Name, HTMLType: "textinput",
				TxType: fitem.Type.String(), Title: fitem.Name,
				Value: value})
		}
	}
	vals := request.ToStringMap()
	vals[`publicKey`] = vals[`public_key_0`]
	txType := "TXNewCitizen"
	return proceedTemplate(c, nCheckCitizen, &checkPage{Data: c.Data, Values: vals,
		Fields: fields, TxType: txType, TxTypeID: utils.TypeInt(txType)})
}
