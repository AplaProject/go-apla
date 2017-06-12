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

	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
)

const aContractInfo = `ajax_contract_info`

// TxSignJSON is a structure for additional signs of transaction
type ContractField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tags string `json:"tags"`
}

// ContractInfo is a structure for the answer of ajax_contract_info ajax request
type ContractInfo struct {
	Fields []ContractField `json:"fields"`
	ID     uint32          `json:"id"`
	Name   string          `json:"name"`
	Error  string          `json:"error"`
}

func init() {
	newPage(aContractInfo, `json`)
}

// AjaxContractInfo returns fields of the contract
func (c *Controller) AjaxContractInfo() interface{} {
	var (
		result ContractInfo
	)
	result.Fields = make([]ContractField, 0)
	result.Name = c.r.FormValue(`name`)
	contract := smart.GetContract(result.Name, uint32(c.SessStateID))
	if contract == nil {
		result.Error = fmt.Sprintf(`there is not %s contract`, result.Name)
	} else {
		result.ID = contract.Block.Info.(*script.ContractInfo).ID
		if contract.Block.Info.(*script.ContractInfo).Tx != nil {
			for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
				result.Fields = append(result.Fields, ContractField{Name: fitem.Name, Type: fitem.Type.String(),
					Tags: fitem.Tags})
			}
		}
	}
	return result
}
