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

package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"

	log "github.com/sirupsen/logrus"
)

type contractField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Optional bool   `json:"optional"`
}

type getContractResult struct {
	ID       uint32          `json:"id"`
	StateID  uint32          `json:"state"`
	Active   bool            `json:"active"`
	TableID  string          `json:"tableid"`
	WalletID string          `json:"walletid"`
	TokenID  string          `json:"tokenid"`
	Address  string          `json:"address"`
	Fields   []contractField `json:"fields"`
	Name     string          `json:"name"`
}

func getContract(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var result getContractResult

	cntname := data.params[`name`].(string)
	contract := smart.VMGetContract(data.vm, cntname, uint32(data.ecosystemId))
	if contract == nil {
		logger.WithFields(log.Fields{"type": consts.ContractError, "contract_name": cntname}).Error("contract name")
		return errorAPI(w, `E_CONTRACT`, http.StatusBadRequest, cntname)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	fields := make([]contractField, 0)
	result = getContractResult{
		ID: info.ID, Name: info.Name, StateID: info.Owner.StateID,
		Active: info.Owner.Active, TableID: converter.Int64ToStr(info.Owner.TableID),
		WalletID: converter.Int64ToStr(info.Owner.WalletID),
		TokenID:  converter.Int64ToStr(info.Owner.TokenID),
		Address:  converter.AddressToString(info.Owner.WalletID),
	}

	if info.Tx != nil {
		for _, fitem := range *info.Tx {
			fields = append(fields, contractField{
				Name:     fitem.Name,
				Type:     getFieldTypeAlias(fitem.Type.String()),
				Optional: fitem.ContainsTag(script.TagOptional),
			})
		}
	}
	result.Fields = fields

	data.result = result
	return nil
}

func getFieldTypeAlias(t string) string {
	var fieldTypeAliases = map[string]string{
		"int64":           "int",
		"float64":         "float",
		"decimal.Decimal": "money",
		"[]uint8":         "bytes",
		"[]interface {}":  "array",
		"types.File":      "file",
	}

	if v, ok := fieldTypeAliases[t]; ok {
		return v
	}
	return t
}
