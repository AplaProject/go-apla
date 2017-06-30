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

import "github.com/EGaaS/go-egaas-mvp/packages/converter"

const aStateCoords = `ajax_state_coords`

// StateDetails is a structure for the answer of ajax_state_coords ajax request
type StateDetails struct {
	Coords string `json:"coords"`
}

func init() {
	newPage(aStateCoords, `json`)
}

// AjaxStateCoords is a controller of ajax_state_coords request
func (c *Controller) AjaxStateCoords() interface{} {

	var stateDetails StateDetails
	stateID := converter.StrToInt64(c.r.FormValue("stateId"))
	stateDetails.Coords, _ = c.Single(`SELECT coords FROM "` + converter.Int64ToStr(stateID) + `_state_details"`).String()
	return stateDetails
}
