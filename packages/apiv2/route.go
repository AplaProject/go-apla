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

package apiv2

import (
	"strings"

	hr "github.com/julienschmidt/httprouter"
)

func methodRoute(route *hr.Router, method, pattern, pars string, handler ...apiHandle) {
	route.Handle(method, `/api/v2/`+pattern, DefaultHandler(processParams(pars), handler...))
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
		methodRoute(route, method, `prepare/`+pattern, pars, authWallet, preHandle)
		if len(pars) > 0 {
			pars = `,` + pars
		}
		methodRoute(route, method, `contract/`+pattern, `?pubkey signature:hex, time:string`+pars, authWallet, handle)
	}
	postTx := func(url string, params string, preHandle, handle apiHandle) {
		anyTx(`POST`, url, params, preHandle, handle)
	}

	route.Handle(`OPTIONS`, `/api/v2/*name`, optionsHandler())

	get(`balance/:wallet`, `?ecosystem:int64`, authWallet, balance)
	get(`contract/:name`, ``, authWallet, getContract)
	get(`contracts`, `?limit ?offset:int64`, authWallet, getContracts)
	get(`ecosystemparam/:name`, `?ecosystem:int64`, authWallet, ecosystemParam)
	get(`ecosystemparams`, `?ecosystem:int64,?names:string`, authWallet, ecosystemParams)
	get(`ecosystems`, ``, authWallet, ecosystems)
	get(`getuid`, ``, getUID)
	get(`list/:name`, `?limit ?offset:int64,?columns:string`, authWallet, list)
	get(`row/:name/:id`, `?columns:string`, authWallet, row)
	get(`table/:name`, ``, authWallet, table)
	get(`tables`, `?limit ?offset:int64`, authWallet, tables)
	get(`txstatus/:hash`, ``, authWallet, txstatus)
	//	get(`smartcontract/:name`, ``, authState, getSmartContract)
	get(`test/:name`, ``, getTest)

	post(`content/page/:name`, ``, authWallet, getPage)
	post(`content/menu/:name`, ``, authWallet, getMenu)
	post(`install`, `?first_load_blockchain_url ?first_block_dir log_level type db_host db_port 
	db_name db_pass db_user:string,?generate_first_block:int64`, install)
	post(`login`, `?pubkey signature:hex,?key_id:string,?ecosystem ?expire:int64`, login)
	postTx(`:name`, `?token_ecosystem:int64,?max_sum ?payover:string`, prepareContract, contract)
	post(`refresh`, `token:string,?expire:int64`, refresh)
	//	postTx(`smartcontract/:name`, ``, txPreSmartContract, txSmartContract)
	post(`signtest/`, `forsign private:string`, signTest)
	post(`test/:name`, ``, getTest)
	post(`content`, `template:string`, jsonContent)

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
			v = strings.TrimSpace(v)
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
