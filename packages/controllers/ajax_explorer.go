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
	"encoding/hex"
	//	"encoding/json"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const AExplorer = `ajax_explorer`

type ExplorerJson struct {
	Data   []map[string]string `json:"data"`
	Latest int64               `json:"latest"`
}

func init() {
	newPage(AExplorer, `json`)
}

func (c *Controller) AjaxExplorer() interface{} {
	/*var (
	      err error
	)*/
	result := ExplorerJson{}
	latest := utils.StrToInt64(c.r.FormValue("latest"))
	if latest > 0 {
		result.Latest, _ = c.Single("select max(id) from block_chain").Int64()
		if result.Latest > latest {
			explorer, err := c.GetAll(`SELECT  b.hash, b.state_id, b.wallet_id, b.time, b.tx, b.id FROM block_chain as b
		where b.id > $1	order by b.id desc limit 30 offset 0`, -1, latest)
			if err == nil {
				for ind := range explorer {
					explorer[ind][`hash`] = hex.EncodeToString([]byte(explorer[ind][`hash`]))
					if len(explorer[ind][`wallet_id`]) > 0 {
						explorer[ind][`wallet_address`] = lib.AddressToString(uint64(utils.StrToInt64(explorer[ind][`wallet_id`])))
					} else {
						explorer[ind][`wallet_address`] = ``
					}

					/*					if explorer[ind][`tx`] == `[]` {
											explorer[ind][`tx_count`] = `0`
										} else {
											var tx []string
											json.Unmarshal([]byte(explorer[ind][`tx`]), &tx)
											if tx != nil && len(tx) > 0 {
												explorer[ind][`tx_count`] = utils.IntToStr(len(tx))
											}
										}*/
				}
				result.Data = explorer
				if explorer != nil && len(explorer) > 0 {
					result.Latest = utils.StrToInt64(explorer[0][`id`])
				}
			}
		}
	}
	if result.Data == nil {
		result.Data = make([]map[string]string, 0)
	}
	return result
}
