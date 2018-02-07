// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package api

import (
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"

	log "github.com/sirupsen/logrus"
)

type contractField struct {
	Name string `json:"name"`
	HTML string `json:"htmltype"`
	Type string `json:"txtype"`
	Tags string `json:"tags"`
}

type getContractResult struct {
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
	result = getContractResult{Name: info.Name, StateID: info.Owner.StateID,
		Active: info.Owner.Active, TableID: converter.Int64ToStr(info.Owner.TableID),
		WalletID: converter.Int64ToStr(info.Owner.WalletID),
		TokenID:  converter.Int64ToStr(info.Owner.TokenID),
		Address:  converter.AddressToString(info.Owner.WalletID)}

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
