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
	"strconv"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const AAddresses = `ajax_addresses`

type AddressJson struct {
	Address []string `json:"address"`
	Error   string   `json:"error"`
}

func init() {
	newPage(AAddresses, `json`)
}

func (c *Controller) AjaxAddresses() interface{} {
	var (
		result AddressJson
		err    error
		req    []map[string]string
	)
	result.Address = make([]string, 0)
	addr := strings.Replace(c.r.FormValue(`address`), `-`, ``, -1)
	state := c.r.FormValue(`state`)
	var request string
	if len(state) == 0 {
		request = `select id from "` + utils.Int64ToStr(c.SessStateId) + `_citizens" where id>=? order by id`
	} else if state == `0` {
		request = `select wallet_id as id from dlt_wallets where wallet_id>=? order by wallet_id`
	} else {
		request = `select id from "` + lib.EscapeName(state) + `_citizens" where id>=? order by id`
	}
	ret, _ := strconv.ParseUint(addr+strings.Repeat(`0`, 20-len(addr)), 10, 64)
	req, err = c.GetAll(request, 7, int64(ret))

	if err != nil {
		result.Error = err.Error()
	} else {
		for _, ireq := range req {
			result.Address = append(result.Address, lib.AddressToString(utils.StrToInt64(ireq[`id`])))
		}
	}
	return result
}
