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

	//	"encoding/json"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

const aExplorer = `ajax_explorer`

// ExplorerJSON is a structure for the answer of ajax_explorer ajax request
type ExplorerJSON struct {
	Data   []map[string]string `json:"data"`
	Latest int64               `json:"latest"`
}

func init() {
	newPage(aExplorer, `json`)
}

// AjaxExplorer is a controller of ajax_explorer request
func (c *Controller) AjaxExplorer() interface{} {
	var blockchain []model.Block
	var err error
	result := ExplorerJSON{}
	latest := converter.StrToInt64(c.r.FormValue("latest"))
	data := make([]map[string]string, 0)

	if latest > 0 {
		block := &model.Block{}
		block.GetMaxBlock()
		result.Latest = block.ID
		if result.Latest > latest {
			blockchain, err = block.GetBlocks(latest, 30)
			if err != nil {
				for _, block := range blockchain {
					data = append(data, block.ToMap())
				}
			}
		}
	}
	result.Data = data
	if data != nil && len(data) > 0 {
		result.Latest = blockchain[0].ID
	}
	return result
}
