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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type interfacePage struct {
	Alert           string
	Lang            map[string]string
	WalletID        int64
	CitizenID       int64
	InterfacePages  []map[string]string
	InterfaceMenu   []map[string]string
	InterfaceBlocks []map[string]string
	Global          string
}

// Interface is a controller for editing pages and menu
func (c *Controller) Interface() (string, error) {

	global := c.r.FormValue("global")
	prefix := c.StateIDStr
	if global == "1" {
		prefix = "global"
	}

	interfacePages, err := c.GetAll(`SELECT * FROM "`+prefix+`_pages" where menu!='0' order by name`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	interfaceBlocks, err := c.GetAll(`SELECT * FROM "`+prefix+`_pages" where menu='0' order by name`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	interfaceMenu, err := c.GetAll(`SELECT * FROM "`+prefix+`_menu" order by name`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("interface", "interface", &interfacePage{
		Alert:           c.Alert,
		Lang:            c.Lang,
		WalletID:        c.SessWalletID,
		CitizenID:       c.SessCitizenID,
		InterfacePages:  interfacePages,
		InterfaceBlocks: interfaceBlocks,
		Global:          global,
		InterfaceMenu:   interfaceMenu})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
