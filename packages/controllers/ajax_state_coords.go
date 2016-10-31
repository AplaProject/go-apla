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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const AStateCoords = `ajax_state_coords`

type StateDetails struct {
	Coords string `json:"coords"`
}

func init() {
	newPage(AStateCoords, `json`)
}

func (c *Controller) AjaxStateCoords() interface{} {

	var stateDetails StateDetails
	stateId := utils.StrToInt64(c.r.FormValue("stateId"))
	stateDetails.Coords, _ = c.Single(`SELECT coords FROM "` + utils.Int64ToStr(stateId) + `_state_details"`).String()
	return stateDetails
}
