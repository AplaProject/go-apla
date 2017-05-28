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
	//	"encoding/json"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const NGenKeys = `gen_keys`

type genKeysPage struct {
	Data      *CommonPage
	Message   string
	Generated int64
	Used      int64
	Available int64
}

func init() {
	newPage(NGenKeys)
}

func (c *Controller) GenKeys() (string, error) {
	govAccount, _ := utils.StateParam(int64(c.SessStateID), `gov_account`)
	if c.SessCitizenID != utils.StrToInt64(govAccount) {
		return ``, fmt.Errorf(`Access denied`)
	}
	generated, err := c.Single(`select count(id) from testnet_keys where id=? and state_id=?`, c.SessCitizenID, c.SessStateID).Int64()
	if err != nil {
		return ``, err
	}
	available, err := c.Single(`select count(id) from testnet_keys where id=? and state_id=? and status=0`, c.SessCitizenID, c.SessStateID).Int64()
	if err != nil {
		return ``, err
	}
	pageData := genKeysPage{Data: c.Data, Generated: generated, Available: available, Used: generated - available}
	return proceedTemplate(c, NGenKeys, &pageData)
}
