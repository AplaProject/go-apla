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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

// AjaxStatesList returns the list of states
func (c *Controller) AjaxStatesList() (string, error) {
	result := make([]map[string]string, 0)
	statesList, err := model.GetAllSystemStatesIDs()
	if err != nil {
		return ``, err
	}
	stateParameter := &model.StateParameter{}
	query := func(id int64, name string) (string, error) {
		stateParameter.SetTablePrefix(converter.Int64ToStr(id))
		err = stateParameter.GetByName(name)
		return stateParameter.Value, err
	}
	for _, id := range statesList {
		if !model.IsNodeState(id, c.r.Host) {
			continue
		}

		stateName, err := query(id, `state_name`)
		if err != nil {
			return ``, err
		}
		stateFlag, err := query(id, `state_flag`)
		if err != nil {
			return ``, err
		}
		stateCoords, err := query(id, `state_coords`)
		if err != nil {
			return ``, err
		}
		iresult := make(map[string]string)
		iresult["state_name"] = stateName
		iresult["state_flag"] = stateFlag
		iresult["state_coords"] = stateCoords
		result = append(result, iresult)
	}
	jsondata, err := json.Marshal(result)
	if err != nil {
		return ``, err
	}
	return string(jsondata), nil
}
