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

package apiv2

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
)

type contractField struct {
	Name string `json:"name"`
	HTML string `json:"htmltype"`
	Type string `json:"txtype"`
	Tags string `json:"tags"`
}

type getContractResult struct {
	Fields   []contractField `json:"fields"`
	Name     string          `json:"name"`
	Active   bool            `json:"active"`
	TableID  int64           `json:"tableid"`
	WalletID int64           `json:"walletid"`
}

func getContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var result getContractResult

	cntname := data.params[`name`].(string)
	contract := smart.GetContract(cntname, uint32(data.state))
	if contract == nil {
		return errorAPI(w, fmt.Sprintf(`there is not %s contract`, cntname), http.StatusBadRequest)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	fields := make([]contractField, 0)
	result = getContractResult{Name: info.Name, Active: info.Active, TableID: info.TableID,
		WalletID: info.WalletID}

	if info.Tx != nil {
		for _, fitem := range *info.Tx {
			field := contractField{Name: fitem.Name, Type: fitem.Type.String(), Tags: fitem.Tags}

			if strings.Contains(fitem.Tags, `hidden`) || strings.Contains(fitem.Tags, `signature`) {
				field.HTML = `hidden`
			} else {
				for _, tag := range []string{`date`, `polymap`, `map`, `image`, `text`, `address`} {
					if strings.Contains(fitem.Tags, tag) {
						field.HTML = tag
						break
					}
				}
				if len(field.HTML) == 0 {
					if fitem.Type.String() == script.Decimal {
						field.HTML = `money`
					} else if fitem.Type.String() == `string` || fitem.Type.String() == `int64` || fitem.Type.String() == `float64` {
						field.HTML = `textinput`
					}
				}
			}
			fields = append(fields, field)
		}
	}
	result.Fields = fields

	data.result = result
	return nil
}
