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

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	hr "github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

func methodRoute(route *hr.Router, method, pattern, pars string, handler ...apiHandle) {
	route.Handle(
		method,
		consts.ApiPath+pattern,
		DefaultHandler(method, pattern, processParams(pars), append([]apiHandle{blockchainUpdatingState}, handler...)...),
	)
}

// Route sets routing pathes
func Route(route *hr.Router) {
	get := func(pattern, params string, handler ...apiHandle) {
		methodRoute(route, `GET`, pattern, params, handler...)
	}
	post := func(pattern, params string, handler ...apiHandle) {
		methodRoute(route, `POST`, pattern, params, handler...)
	}
	contractHandlers := &contractHandlers{
		requests:      tx.NewRequestBuffer(consts.TxRequestExpire),
		multiRequests: tx.NewMultiRequestBuffer(consts.TxRequestExpire),
	}

	route.Handle(`OPTIONS`, consts.ApiPath+`*name`, optionsHandler())
	route.Handle(`GET`, consts.ApiPath+`data/:table/:id/:column/:hash`, dataHandler())

	get(`appparam/:appid/:name`, `?ecosystem:int64`, authWallet, appParam)
	get(`appparams/:appid`, `?ecosystem:int64,?names:string`, authWallet, appParams)
	get(`balance/:wallet`, `?ecosystem:int64`, authWallet, balance)
	get(`contract/:name`, ``, authWallet, getContract)
	get(`contracts`, `?limit ?offset:int64`, authWallet, getContracts)
	get(`ecosystemparam/:name`, `?ecosystem:int64`, authWallet, ecosystemParam)
	get(`ecosystemparams`, `?ecosystem:int64,?names:string`, authWallet, ecosystemParams)
	get(`ecosystems`, ``, authWallet, ecosystems)
	get(`getuid`, ``, getUID)
	get(`list/:name`, `?limit ?offset:int64,?columns:string`, authWallet, list)
	get(`row/:name/:id`, `?columns:string`, authWallet, row)
	get(`interface/page/:name`, ``, authWallet, getPageRow)
	get(`interface/menu/:name`, ``, authWallet, getMenuRow)
	get(`interface/block/:name`, ``, authWallet, getBlockInterfaceRow)
	get(`systemparams`, `?names:string`, authWallet, systemParams)
	get(`table/:name`, ``, authWallet, table)
	get(`tables`, `?limit ?offset:int64`, authWallet, tables)
	get(`txstatus/:hash`, ``, authWallet, txstatus)
	get(`test/:name`, ``, getTest)
	get(`history/:table/:id`, ``, authWallet, getHistory)
	get(`block/:id`, ``, getBlockInfo)
	get(`maxblockid`, ``, getMaxBlockID)
	get(`version`, ``, getVersion)
	get(`avatar/:ecosystem/:member`, ``, getAvatar)
	get(`config/:option`, ``, getConfigOption)
	get("ecosystemname", "?id:int64", getEcosystemName)
	post(`content/source/:name`, ``, authWallet, getSource)
	post(`content/page/:name`, `?lang:string`, authWallet, getPage)
	post(`content/menu/:name`, `?lang:string`, authWallet, getMenu)
	post(`content/hash/:name`, ``, getPageHash)
	post(`vde/create`, ``, authWallet, vdeCreate)
	post(`login`, `?pubkey signature:hex,?key_id ?mobile:string,?ecosystem ?expire ?role_id:int64`, login)
	post(`prepare/:name`, `?token_ecosystem:int64,?max_sum ?payover:string`, authWallet, contractHandlers.prepareContract)
	post(`prepareMultiple`, `data:string`, authWallet, contractHandlers.prepareMultipleContract)
	post(`txstatusMultiple`, `data:string`, authWallet, txstatusMulti)
	post(`contract/:request_id`, `?pubkey signature:hex, time:string, ?token_ecosystem:int64,?max_sum ?payover:string`, authWallet, blockchainUpdatingState, contractHandlers.contract)
	post(`contractMultiple/:request_id`, `data:string`, authWallet, blockchainUpdatingState, contractHandlers.contractMulti)
	post(`refresh`, `token:string,?expire:int64`, refresh)
	post(`test/:name`, ``, getTest)
	post(`content`, `template ?source:string`, jsonContent)
	post(`updnotificator`, `ids:string`, updateNotificator)

	methodRoute(route, `POST`, `node/:name`, `?token_ecosystem:int64,?max_sum ?payover:string`, contractHandlers.nodeContract)
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
