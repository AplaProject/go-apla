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

// Route sets routing pathes
func Route(route *hr.Router) {
	anyMethod := func(method, pattern, pars string, handler ...apiHandle) {
		route.Handle(method, `/api/v1/`+pattern, DefaultHandler(processParams(pars), handler...))
	}
	get := func(pattern string, handler ...apiHandle) {
		anyMethod(`GET`, pattern, ``, handler...)
	}
	post := func(pattern, params string, handler ...apiHandle) {
		anyMethod(`POST`, pattern, params, handler...)
	}
	put := func(pattern, params string, handler ...apiHandle) {
		anyMethod(`PUT`, pattern, params, handler...)
	}
	getGlobal := func(url string, handler apiHandle) {
		get(url+`/:global`, authState, handler)
		get(url, authState, handler)
	}
	postTx := func(url string, params string, preHandle, handle apiHandle) {
		post(`prepare/`+url, params, authState, preHandle)
		post(url, `signature:hex, time:string, `+params, authState, handle)
	}
	putTx := func(url string, params string, preHandle, handle apiHandle) {
		put(`prepare/`+url, params, authState, preHandle)
		put(url, `signature:hex, time:string, `+params, authState, handle)
	}

	get(`balance/:wallet`, authWallet, balance)
	get(`getuid`, getUID)
	get(`txstatus/:hash`, authWallet, txstatus)
	getGlobal(`content/page/:page`, contentPage)
	getGlobal(`content/menu/:name`, contentMenu)
	getGlobal(`menu/:name`, getMenu)

	post(`login`, `pubkey signature:hex,?state:int64`, login)
	//	post(`prepare/menu`, `name value conditions:string, global:int64`, authState, prePostMenu)
	//	post(`menu`, `signature:hex, time name value conditions:string, global:int64`, authState, postMenu)
	postTx(`menu`, `name value conditions:string, global:int64`, prePostMenu, postMenu)
	post(`prepare/sendegs`, `recipient amount commission ?comment:string`, authWallet, preSendEGS)
	post(`sendegs`, `pubkey signature:hex, time recipient amount commission ?comment:string`, authWallet, sendEGS)

	//	put(`prepare/menu/:name`, `value conditions:string`, authState, prePutMenu)
	//	put(`menu/:name`, `signature:hex, time value conditions:string`, authState, putMenu)
	putTx(`menu/:name`, `value conditions:string, global:int64`, prePutMenu, putMenu)
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
