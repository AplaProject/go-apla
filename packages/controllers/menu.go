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

import "regexp"

//"fmt"

const NMenu = `menu`

type menuPage struct {
	Data *CommonPage
	Menu string
}

func init() {
	newPage(NMenu)
}

func (c *Controller) Menu() (string, error) {
	var err error
	menu := ""
	if c.StateIdStr != "" {
		menu, err = c.Single(`SELECT value FROM "`+c.StateIdStr+`_menu" WHERE name = ?`, "menu_default").String()
		if err != nil {
			return "", err
		}

		qrx := regexp.MustCompile(`(?is)\[([\w\s]*)\]\(([\w\s]*)\)`)
		menu = qrx.ReplaceAllString(menu, "<li><a href='#' onclick=\"load_template('$2'); HideMenu();\"><span>$1</span></a></li>")
		qrx = regexp.MustCompile(`(?is)\[([\w\s]*)\]\(sys.([\w\s]*)\)`)
		menu = qrx.ReplaceAllString(menu, "<li><a href='#' onclick=\"load_page('$2'); HideMenu();\"><span>$1</span></a></li>")

	}
	return proceedTemplate(c, NMenu, &menuPage{Data: c.Data, Menu: menu})
}
