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
	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
)

const aGetCnt = `ajax_get_cnt`

// GetCntJSON is a structure for the answer of ajax_citizen_fields ajax request
type GetCntJSON struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

func init() {
	newPage(aGetCnt, `json`)
}

// AjaxGetCnt is a controller of ajax_get_cnt request
func (c *Controller) AjaxGetCnt() interface{} {
	var result GetCntJSON

	id, err := strconv.ParseInt(c.r.FormValue(`id`), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrToIntError, c.r.FormValue("id"))
	}
	if id > 0 {
		contract := smart.GetContractByID(int32(id))
		if contract != nil {
			result.Name = contract.Name
		} else {
			result.Name = fmt.Sprintf(`Unknown contract %d`, id)
		}
	}
	return result
}
