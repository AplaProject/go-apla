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
	"encoding/json"
)


func (c *Controller) AjaxStatesList() (string, error) {

	result := make(map[string]map[string]string)
	data,err := c.GetList(`SELECT id FROM system_states`).String()
	if err!=nil {
		return ``, err
	}
	for _, id := range data {
		state_name, err := c.Single(`SELECT value FROM "`+id+`_state_parameters" WHERE name = 'state_name'`).String()
		if err!=nil {
			return ``, err
		}
		state_flag,err := c.Single(`SELECT value FROM "`+id+`_state_parameters" WHERE name = 'state_flag'`).String()
		if err!=nil {
			return ``, err
		}
		state_coords,err := c.Single(`SELECT value FROM "`+id+`_state_parameters" WHERE name = 'state_coords'`).String()
		if err!=nil {
			return ``, err
		}
		result[id] = make(map[string]string)
		result[id]["state_name"] = state_name
		result[id]["state_flag"] = state_flag
		result[id]["state_coords"] = state_coords

	}
	jsondata, _ := json.Marshal(result)
	return string(jsondata), nil
}
