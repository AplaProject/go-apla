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
	"regexp"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/textproc"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

//"fmt"

const NMenu = `menu`

type menuPage struct {
	Data          *CommonPage
	Menu          string
	MainMenu      string
	CanCitizen    bool
	StateName     string
	StateFlag     string
	CitizenName   string
	CitizenAvatar string
	UpdVer        string
	Btc           string
}

func init() {
	newPage(NMenu)
}

func ReplaceMenu(menu string) string {
	qrx := regexp.MustCompile(`(?is)\[([\w\s]*)\]\(glob.([\w\s]*)\){?([\w\d\s""'',:]*)?}?`)
	menu = qrx.ReplaceAllString(menu, "<li class='citizen_$2'><a href='#' onclick=\"load_template('$2',{global:1, $3}); HideMenu();\"><span>$1</span></a></li>")
	qrx = regexp.MustCompile(`(?is)\[([\w\s]*)\]\(([\w\s]*)\){?([\w\d\s"",:]*)?}?`)
	menu = qrx.ReplaceAllString(menu, "<li class='citizen_$2'><a href='#' onclick=\"load_template('$2',{$3}); HideMenu();\"><span>$1</span></a></li>")
	qrx = regexp.MustCompile(`(?is)\[([\w\s]*)\]\(sys.([\w\s]*)\){?([\w\d\s"",:]*)?}?`)
	return qrx.ReplaceAllString(menu, "<li class='citizen_$2'><a href='#' onclick=\"load_page('$2', {$3}); HideMenu();\"><span>$1</span></a></li>")
}

func (c *Controller) Menu() (string, error) {
	var (
		err                                                                  error
		updver, menu, main, stateName, stateFlag, citizenName, citizenAvatar string
	)

	if strings.HasPrefix(c.r.Host, `localhost`) {
		updinfo, err := utils.GetUpdVerAndUrl(consts.UPD_AND_VER_URL)
		if err == nil && updinfo != nil {
			updver = updinfo.Version
		}
	}

	canCitizen, _ := c.Single(`SELECT count(id) FROM system_states`).Int64()
	if c.StateIdStr != "" {
		main, err = c.Single(`SELECT value FROM "`+c.StateIdStr+`_menu" WHERE name = ?`, "xmain_menu").String()
		if err != nil {
			return "", err
		}
		if len(main) > 0 {
			params := make(map[string]string)
			params[`state_id`] = c.StateIdStr
			params[`accept_lang`] = c.r.Header.Get(`Accept-Language`)
			main = utils.LangMacro(textproc.Process(main, &params), utils.StrToInt(c.StateIdStr), params[`accept_lang`])
		} else {
			menu, err = c.Single(`SELECT value FROM "`+c.StateIdStr+`_menu" WHERE name = ?`, "menu_default").String()
			if err != nil {
				return "", err
			}
		}

		stateName, err = c.Single(`SELECT value FROM "`+c.StateIdStr+`_state_parameters" WHERE name = ?`, "state_name").String()
		if err != nil {
			return "", err
		}
		stateFlag, err = c.Single(`SELECT value FROM "`+c.StateIdStr+`_state_parameters" WHERE name = ?`, "state_flag").String()
		if err != nil {
			return "", err
		}

		citizenName, err = c.Single(`SELECT name FROM "`+c.StateIdStr+`_citizens" WHERE id = ?`, c.SessCitizenId).String()
		if err != nil {
			log.Error("%v", err)
		}

		citizenAvatar, err = c.Single(`SELECT avatar FROM "`+c.StateIdStr+`_citizens" WHERE id = ?`, c.SessCitizenId).String()
		if err != nil {
			log.Error("%v", err)
		}
		menu = ReplaceMenu(menu)
	}
	return proceedTemplate(c, NMenu, &menuPage{Data: c.Data, Menu: menu, MainMenu: main, CanCitizen: canCitizen > 0,
		StateName: stateName, StateFlag: stateFlag, CitizenName: citizenName,
		CitizenAvatar: citizenAvatar, UpdVer: updver, Btc: GetBtc()})
}
