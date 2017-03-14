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
	"io/ioutil"
	"path/filepath"
	"strconv"
)

type loginECDSAPage struct {
	Lang  map[string]string
	Title string
	Key   string
	//	States     map[string]string
	States      string
	State       int64
	OneCountry  int64
	PrivCountry bool
	Import      bool
	Private     string
}

func (c *Controller) LoginECDSA() (string, error) {
	var err error
	var private []byte
	if c.ConfigIni["public_node"] != "1" {
		private, _ = ioutil.ReadFile(filepath.Join(*utils.Dir, `PrivateKey`))
	}

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
	key := c.r.FormValue("key")
	pkey := c.r.FormValue("pkey")
	state := c.r.FormValue("state")
	if len(key) > 0 || len(pkey) > 0 {
		c.Logout()
	}
	if len(pkey) > 0 {
		private = []byte(pkey)
	}
	var state_id int64
	if len(state) > 0 {
		state_id, err = strconv.ParseInt(state, 10, 64)
		if err != nil {
			list, err := utils.DB.GetAllTables()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if utils.InSliceString(`global_states_list`, list) {
				state_id, err = c.Single("select state_id from global_states_list where state_name=?", state).Int64()
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			}
		}
	}
	TemplateStr, err := makeTemplate("login", "loginECDSA", &loginECDSAPage{
		Lang:        c.Lang,
		Title:       "Login",
		States:      states,
		State:       state_id,
		Key:         key,
		Import:      len(pkey) > 0,
		OneCountry:  utils.OneCountry,
		PrivCountry: utils.PrivCountry,
		Private:     string(private),
		/*		MyWalletData:          MyWalletData,
				Title:                 "modalAnonym",
		*/})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
