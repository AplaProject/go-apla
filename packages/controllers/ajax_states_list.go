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
		log.Errorf("can't get all stated ids: %s", err)
		return ``, err
	}
	log.Debugf("states list: %+v", statesList)

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

		stateParams := make(map[string]string)
		for _, paramName := range []string{"state_name", "state_flag", "state_coords"} {
			param, err := query(id, paramName)
			if err != nil {
				// TODO: ???? return "", err
			}
			stateParams[paramName] = param
		}
		result = append(result, stateParams)
	}

	jsondata, err := json.Marshal(result)
	if err != nil {
		return ``, err
	}
	return string(jsondata), nil
}
