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
	//	"io/ioutil"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	//	"path/filepath"
)

type loginECDSAPage struct {
	Lang  map[string]string
	Title string
	//	States     map[string]string
	States      string
	OneCountry  int64
	PrivCountry bool
	//	Private string
}

func (c *Controller) LoginECDSA() (string, error) {

	//	var private []byte
	//	private, _ = ioutil.ReadFile(filepath.Join(*utils.Dir, `PrivateKey`))

	/*	states := make(map[string]string)
		data, err := c.GetList(`SELECT id FROM system_states`).String()
		if err != nil {
			return ``, err
		}
		for _, id := range data {
			state_name, err := c.Single(`SELECT value FROM "` + id + `_state_parameters" WHERE name = 'state_name'`).String()
			if err != nil {
				return ``, err
			}
			states[id] = state_name
		}*/
	states, _ := c.AjaxStatesList()

	TemplateStr, err := makeTemplate("login", "loginECDSA", &loginECDSAPage{
		Lang:        c.Lang,
		Title:       "Login",
		States:      states,
		OneCountry:  utils.OneCountry,
		PrivCountry: utils.PrivCountry,
		//	Private: string(private),
		/*		MyWalletData:          MyWalletData,
				Title:                 "modalAnonym",
		*/})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
