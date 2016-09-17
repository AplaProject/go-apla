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

//	"fmt"

//	"github.com/DayLightProject/go-daylight/packages/utils"

const NCheckCitizen = `check_citizen_status`

type checkPage struct {
	Data   *CommonPage
	Values map[string]string
	Fields []FieldInfo
}

func init() {
	newPage(NCheckCitizen)
}

func (c *Controller) CheckCitizenStatus() (string, error) {
	var fields []FieldInfo
	pref := `ds`
	field, err := c.Single(`SELECT value FROM ` + pref + `_state_settings where parameter='citizen_fields'`).String()
	if err != nil {
		return ``, err
	}
	if err = json.Unmarshal([]byte(field), &fields); err != nil {
		return ``, err
	}
	vals, err := c.OneRow(`select * from ` + pref + `_citizens_requests_private where approved=0 order by id`).String()
	if err != nil {
		return ``, err
	}
	return proceedTemplate(c, NCheckCitizen, &checkPage{Data: c.Data, Values: vals, Fields: fields})
}
