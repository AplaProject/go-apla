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
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/modes"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/daemons"
	hr "github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

func MethodRoute(route *hr.Router, method, pattern, pars string, mode modes.NodeMode, handler ...ApiHandle) {
	route.Handle(method, consts.ApiPath+pattern, DefaultHandler(method, pattern, mode, processParams(pars), handler...))
}

func setRoute(route *hr.Router, path string, handle func(http.ResponseWriter, *http.Request), methods ...string) {
	for _, method := range methods {
		route.HandlerFunc(method, path, handle)
	}
}

// CreateDefaultRouter returns common api
func CreateDefaultRouter(mode modes.NodeMode) *hr.Router {
	router := hr.New()
	setRoute(router, `/monitoring`, daemons.Monitoring, `GET`)
	AddCommonRoutes(router, mode)

	router.Handler(`GET`, consts.WellKnownRoute, http.FileServer(http.Dir(*conf.TLS)))
	return router
}

// AddCommonRoutes add common handlers to router
func AddCommonRoutes(router *hr.Router, mode modes.NodeMode) {

	get := func(pattern, params string, handler ...ApiHandle) {
		MethodRoute(router, `GET`, pattern, params, mode, handler...)
	}
	post := func(pattern, params string, handler ...ApiHandle) {
		MethodRoute(router, `POST`, pattern, params, mode, handler...)
	}
	anyTx := func(method, pattern, pars string, preHandle, handle ApiHandle) {
		MethodRoute(router, method, `prepare/`+pattern, pars, mode, authWallet, preHandle)
		if len(pars) > 0 {
			pars = `,` + pars
		}
		MethodRoute(router, method, `contract/`+pattern, `?pubkey signature:hex, time:string`+pars, mode, authWallet, handle)
	}
	postTx := func(url string, params string, preHandle, handle ApiHandle) {
		anyTx(`POST`, url, params, preHandle, handle)
	}

	router.Handle(`OPTIONS`, consts.ApiPath+`*name`, optionsHandler())
	router.Handle(`GET`, consts.ApiPath+`data/:table/:id/:column/:hash`, dataHandler())

	get(`balance/:wallet`, `?ecosystem:int64`, authWallet, balance)
	get(`contract/:name`, ``, authWallet, getContract)
	get(`contracts`, `?limit ?offset:int64`, authWallet, getContracts)

	get(`getuid`, ``, getUID)
	get(`list/:name`, `?limit ?offset:int64,?columns:string`, authWallet, list)
	get(`row/:name/:id`, `?columns:string`, authWallet, row)
	get(`systemparams`, `?names:string`, authWallet, systemParams)
	get(`table/:name`, ``, authWallet, table)
	get(`tables`, `?limit ?offset:int64`, authWallet, tables)

	get(`test/:name`, ``, getTest)
	get(`history/:table/:id`, ``, authWallet, getHistory)

	post(`content/source/:name`, ``, authWallet, getSource)
	post(`content/page/:name`, `?lang:string`, authWallet, getPage)
	post(`content/menu/:name`, `?lang:string`, authWallet, getMenu)
	post(`content/hash/:name`, ``, authWallet, getPageHash)
	post(`install`, `?first_load_blockchain_url ?first_block_dir log_level type db_host db_port 
	db_name db_pass db_user ?centrifugo_url ?centrifugo_secret:string,?generate_first_block:int64`, doInstall)
	// post(`vde/create`, ``, authWallet, vdeCreate)
	post(`login`, `?pubkey signature:hex,?key_id:string,?ecosystem ?expire:int64`, login)
	postTx(`:name`, `?token_ecosystem:int64,?max_sum ?payover:string`, prepareContract, contract)
	post(`refresh`, `token:string,?expire:int64`, refresh)
	post(`signtest/`, `forsign private:string`, signTest)
	post(`test/:name`, ``, getTest)
	post(`content`, `template:string`, jsonContent)

	MethodRoute(router, `POST`, `node/:name`, `?token_ecosystem:int64,?max_sum ?payover:string`, mode, nodeContract)

}

// AddBlockChainRoutes add specific routes to router
func AddBlockChainRoutes(router *hr.Router, mode modes.NodeMode) {

	get := func(pattern, params string, handler ...ApiHandle) {
		MethodRoute(router, `GET`, pattern, params, mode, handler...)
	}

	get(`ecosystemparam/:name`, `?ecosystem:int64`, authWallet, ecosystemParam)
	get(`ecosystemparams`, `?ecosystem:int64,?names:string`, authWallet, ecosystemParams)
	get(`ecosystems`, ``, authWallet, ecosystems)
	get(`txstatus/:hash`, ``, authWallet, txstatus)
	get(`block/:id`, ``, getBlockInfo)
	get(`maxblockid`, ``, getMaxBlockID)
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
			log.WithFields(log.Fields{"type": consts.RouteError, "parameter": par}).Fatal("Incorrect api route parameters")
		}
		switch types[1] {
		case `hex`:
			vtype = pHex
		case `string`:
			vtype = pString
		case `int64`:
			vtype = pInt64
		default:
			log.WithFields(log.Fields{"type": consts.RouteError, "parameter": par}).Fatal("Unknown type of api route parameter")
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
					log.WithFields(log.Fields{"type": consts.RouteError, "parameter": par}).Fatal("Incorrect name of api route parameter")
				}
			} else {
				params[v] = vtype
			}
		}
	}
	return
}
