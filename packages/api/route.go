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

package api

import (
	"strings"

	hr "github.com/julienschmidt/httprouter"
)

func methodRoute(route *hr.Router, method, pattern, pars string, handler ...apiHandle) {
	route.Handle(method, `/api/v1/`+pattern, DefaultHandler(processParams(pars), handler...))
}

// Route sets routing pathes
func Route(route *hr.Router) {
	get := func(pattern, params string, handler ...apiHandle) {
		methodRoute(route, `GET`, pattern, params, handler...)
	}
	post := func(pattern, params string, handler ...apiHandle) {
		methodRoute(route, `POST`, pattern, params, handler...)
	}
	anyTx := func(method, pattern, pars string, preHandle, handle apiHandle) {
		methodRoute(route, method, `prepare/`+pattern, pars, authState, preHandle)
		if len(pars) > 0 {
			pars = `,` + pars
		}
		methodRoute(route, method, pattern, `?pubkey signature:hex, time:string`+pars, authState, handle)
	}
	postTx := func(url string, params string, preHandle, handle apiHandle) {
		anyTx(`POST`, url, params, preHandle, handle)
	}
	putTx := func(url string, params string, preHandle, handle apiHandle) {
		anyTx(`PUT`, url, params, preHandle, handle)
	}

	get(`balance/:wallet`, ``, authWallet, balance)
	get(`getuid`, ``, getUID)
	get(`txstatus/:hash`, ``, authWallet, txstatus)
	get(`smartcontract/:name`, ``, authState, getSmartContract)
	get(`test/:name`, ``, getTest)
	get(`content/page/:page`, `?global:int64`, authState, contentPage)
	get(`content/menu/:name`, `?global:int64`, authState, contentMenu)
	get(`lang/:name`, ``, authState, getLang)
	get(`langlist`, `?limit ?offset:int64`, authState, langList)
	get(`menu/:name`, `?global:int64`, authState, getMenu)
	get(`menulist`, `?limit ?offset ?global:int64`, authState, menuList)
	get(`page/:name`, `?global:int64`, authState, getPage)
	get(`pagelist`, `?limit ?offset ?global:int64`, authState, pageList)
	get(`statelist`, `?limit ?offset:int64`, authState, stateList)
	get(`stateparams/:name`, ``, authState, getStateParams)
	get(`stateparamslist`, `?limit ?offset:int64`, authState, stateParamsList)
	get(`table/:name`, `?global:int64`, authState, getTable)
	get(`tables`, `?limit ?offset ?global:int64`, authState, tableList)
	get(`contract/:id`, `?global:int64`, authState, getContract)
	get(`contractlist`, `?limit ?offset ?global:int64`, authState, contractList)

	post(`login`, `pubkey signature:hex,?state:int64`, login)
	postTx(`column/:table`, `name type permissions:string, index:int64`, txPreNewColumn, txNewColumn)
	postTx(`lang`, `name trans:string`, txPreNewLang, txLang)
	postTx(`menu`, `name value conditions:string, global:int64`, txPreNewMenu, txMenu)
	postTx(`newstate`, `name currency:string`, txPreNewState, txNewState)
	postTx(`table`, `name columns:string, global:int64`, txPreNewTable, txNewTable)
	postTx(`page`, `name menu value conditions:string, global:int64`, txPreNewPage, txPage)
	postTx(`stateparams`, `name value conditions:string`, txPreNewStateParams, txStateParams)
	postTx(`contract`, `name value conditions ?wallet:string, global:int64`, txPreNewContract, txContract)
	postTx(`smartcontract/:name`, ``, txPreSmartContract, txSmartContract)
	post(`prepare/sendegs`, `recipient amount commission ?comment:string`, authWallet, preSendEGS)
	post(`sendegs`, `pubkey signature:hex, time recipient amount commission ?comment:string`, authWallet, sendEGS)

	putTx(`activatecontract/:id`, `?global:int64`, txPreActivateContract, txActivateContract)
	putTx(`contract/:id`, `value conditions:string, global:int64`, txPreEditContract, txContract)
	putTx(`lang/:name`, `trans:string`, txPreEditLang, txLang)
	putTx(`menu/:name`, `value conditions:string, global:int64`, txPreEditMenu, txMenu)
	putTx(`appendmenu/:name`, `value:string, global:int64`, txPreAppendMenu, txAppendMenu)
	putTx(`page/:name`, `menu value conditions:string, global:int64`, txPreEditPage, txPage)
	putTx(`appendpage/:name`, `value:string, global:int64`, txPreAppendPage, txAppendPage)
	putTx(`stateparams/:name`, `value conditions:string`, txPreEditStateParams, txStateParams)
	putTx(`table/:name`, `insert new_column general_update:string`, txPreEditTable, txEditTable)
	putTx(`column/:table/:name`, `permissions:string`, txPreEditColumn, txEditColumn)
}

func processParams(input string) (params map[string]int) {
	if len(input) == 0 {
		return
	}
	params = make(map[string]int)
	for _, par := range strings.Split(input, `,`) {
		var vtype int
		types := strings.Split(par, `:`)
		if len(types) != 2 {
			log.Fatalf(`Incorrect api route parameters: "%s"`, par)
		}
		switch types[1] {
		case `hex`:
			vtype = pHex
		case `string`:
			vtype = pString
		case `int64`:
			vtype = pInt64
		default:
			log.Fatalf(`Unknown type of api route parameter: "%s"`, par)
		}
		vars := strings.Split(types[0], ` `)
		for _, v := range vars {
			if len(v) == 0 {
				continue
			}
			if v[0] == '?' {
				if len(v) > 1 {
					params[v[1:]] = vtype | pOptional
				} else {
					log.Fatalf(`Incorrect name of api route parameter: "%s"`, par)
				}
			} else {
				params[v] = vtype
			}
		}
	}
	return
}
