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
	//	"fmt"
	"strconv"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

const aAddresses = `ajax_addresses`

// AddressJSON is a structure of the ajax_adresses ajax request
type AddressJSON struct {
	Address []string `json:"address"`
	Error   string   `json:"error"`
}

func init() {
	newPage(aAddresses, `json`)
}

// AjaxAddresses is a controller of ajax_adresses request
func (c *Controller) AjaxAddresses() interface{} {
	var (
		result AddressJSON
		err    error
		req    []map[string]string
	)
	result.Address = make([]string, 0)
	walletAddress := strings.Replace(c.r.FormValue(`address`), `-`, ``, -1)
	state := c.r.FormValue(`state`)
	startCitizenID, _ := strconv.ParseInt(walletAddress+strings.Repeat(`0`, 20-len(walletAddress)), 10, 64)
	var citizens []model.Citizens
	var dltWallets []model.DltWallets

	if len(state) == 0 {
		citizens, err = model.GetAllCitizensWhereIdMoreThan(c.SessStateID, startCitizenID, 7)
		if err != nil {
			result.Error = err.Error()
		} else {
			for _, citizen := range citizens {
				result.Address = append(result.Address, converter.AddressToString(citizen.ID))
			}
		}
	} else if state == `0` {
		dltWallets, err = model.GetWallets(startCitizenID, 7)
		if err != nil {
			result.Error = err.Error()
		} else {
			for _, wallet := range dltWallets {
				result.Address = append(result.Address, converter.AddressToString(wallet.WalletID))
			}
		}
	} else {
		tablePrefix, err := strconv.ParseInt(converter.EscapeName(state), 10, 64)
		if err != nil {
			result.Error = err.Error()
		}
		citizens, err = model.GetAllCitizensWhereIdMoreThan(tablePrefix, startCitizenID, 7)
		if err != nil {
			result.Error = err.Error()
		} else {
			for _, ireq := range req {
				result.Address = append(result.Address, converter.AddressToString(converter.StrToInt64(ireq[`id`])))
			}
		}
	}

	return result
}
