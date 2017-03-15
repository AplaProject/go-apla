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
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const AGetCnt = `ajax_get_cnt`

type GetCntJson struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

func init() {
	newPage(AGetCnt, `json`)
}

func (c *Controller) AjaxGetCnt() interface{} {
	var result GetCntJson

	id := utils.StrToInt64(c.r.FormValue(`id`))
	if id > 0 {
		contract := smart.GetContractById(int32(id))
		if contract != nil {
			result.Name = contract.Name
		} else {
			result.Name = fmt.Sprintf(`Unknown contract %d`, id)
		}
	}
	return result
}
