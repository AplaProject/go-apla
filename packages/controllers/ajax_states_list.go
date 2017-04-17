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
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func (c *Controller) AjaxStatesList() (string, error) {

	result := make(map[string]map[string]string)
	data, err := c.GetList(`SELECT id FROM system_states`).String()
	if err != nil {
		return ``, err
	}
	query := func(id string, name string) (string, error) {
		return c.Single(fmt.Sprintf(`SELECT value FROM "%d_state_parameters" WHERE name = ?`, id), name).String()
	}
	for _, id := range data {
		if !c.IsNodeState(utils.StrToInt64(id), c.r.Host) {
			continue
		}

		state_name, err := query(id, `state_name`)
		if err != nil {
			return ``, err
		}
		state_flag, err := query(id, `state_flag`)
		if err != nil {
			return ``, err
		}
		state_coords, err := query(id, `state_coords`)
		if err != nil {
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
