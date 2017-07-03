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
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// NewMenu is a controller for creating a new menu
func (c *Controller) NewMenu() (string, error) {

	txType := "NewMenu"
	timeNow := time.Now().Unix()

	global := c.r.FormValue("global")
	if global != "1" {
		global = "0"
	}

	TemplateStr, err := makeTemplate("edit_menu", "editMenu", &editMenuPage{
		Alert:     c.Alert,
		Lang:      c.Lang,
		Global:    global,
		WalletID:  c.SessWalletID,
		CitizenID: c.SessCitizenID,
		TimeNow:   timeNow,
		TxType:    txType,
		TxTypeID:  utils.TypeInt(txType),
		StateID:   c.SessStateID,
		DataMenu:  map[string]string{`conditions`: "ContractConditions(`MainCondition`)"}})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
