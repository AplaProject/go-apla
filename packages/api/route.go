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

func optionalRoute(route *hr.Router, method, pattern, pars string, handler ...apiHandle) {
	var url string
	path := strings.Split(pattern, `/:?`)
	for _, item := range path {
		if len(url) > 0 {
			url += `/:`
		}
		url += item
		methodRoute(route, method, url, pars, handler...)
	}
}

// Route sets routing pathes
func Route(route *hr.Router) {
	get := func(pattern string, handler ...apiHandle) {
		optionalRoute(route, `GET`, pattern, ``, handler...)
	}
	post := func(pattern, params string, handler ...apiHandle) {
		methodRoute(route, `POST`, pattern, params, handler...)
	}
	/*	put := func(pattern, params string, handler ...apiHandle) {
		anyMethod(`PUT`, pattern, params, handler...)
	}*/
	getOptional := func(url string, handler apiHandle) {
		optionalRoute(route, `GET`, url, ``, authState, handler)
	}
	anyTx := func(method, pattern, pars string, preHandle, handle apiHandle) {
		optionalRoute(route, method, `prepare/`+pattern, pars, authState, preHandle)
		if len(pars) > 0 {
			pars = `,` + pars
		}
		optionalRoute(route, method, pattern, `signature:hex, time:string`+pars, authState, handle)
	}
	postTx := func(url string, params string, preHandle, handle apiHandle) {
		anyTx(`POST`, url, params, preHandle, handle)
	}
	putTx := func(url string, params string, preHandle, handle apiHandle) {
		anyTx(`PUT`, url, params, preHandle, handle)
	}

	get(`balance/:wallet`, authWallet, balance)
	get(`getuid`, getUID)
	get(`txstatus/:hash`, authWallet, txstatus)
	getOptional(`content/page/:page/:?global`, contentPage)
	getOptional(`content/menu/:name/:?global`, contentMenu)
	getOptional(`menu/:name/:?global`, getMenu)
	getOptional(`page/:name/:?global`, getPage)
	getOptional(`contract/:id/:?global`, getContract)
	getOptional(`contractlist/:?limit/:?offset/:?global`, contractList)

	post(`login`, `pubkey signature:hex,?state:int64`, login)
	postTx(`menu`, `name value conditions:string, global:int64`, txPreMenu, txMenu)
	postTx(`page`, `name menu value conditions:string, global:int64`, txPrePage, txPage)
	postTx(`contract`, `name value conditions:string, ?wallet global:int64`, txPreContract, txContract)
	post(`prepare/sendegs`, `recipient amount commission ?comment:string`, authWallet, preSendEGS)
	post(`sendegs`, `pubkey signature:hex, time recipient amount commission ?comment:string`, authWallet, sendEGS)

	putTx(`activatecontract/:id/:?global`, ``, txPreActivateContract, txActivateContract)
	putTx(`contract/:id`, `value conditions:string, global:int64`, txPreContract, txContract)
	putTx(`menu/:name`, `value conditions:string, global:int64`, txPreMenu, txMenu)
	putTx(`page/:name`, `menu value conditions:string, global:int64`, txPrePage, txPage)
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
