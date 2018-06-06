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
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/gorilla/mux"
)

// Route sets routing paths
func Route(router *mux.Router) {
	// TODO: cors

	router.StrictSlash(true)
	router.Use(LoggerMiddleware)

	// api router with prefix path
	api := router.PathPrefix("/api/v2").Subrouter()
	api.Use(NodeStateMiddleware, TokenMiddleware, ClientMiddleware)

	contractHandlers := &contractHandlers{
		requests:      tx.NewRequestBuffer(consts.TxRequestExpire),
		multiRequests: tx.NewMultiRequestBuffer(consts.TxRequestExpire),
	}

	api.HandleFunc("/data/{table}/{id}/{column}/{hash}", dataHandler).Methods("GET")
	api.HandleFunc("/data/{prefix}_binaries/{id}/data/{hash}", binaryHandler).Methods("GET")
	api.HandleFunc("/appparam/{id}/{name}", AuthRequire(appParamHandler)).Methods("GET")        // get(`appparam/:appid/:name`, `?ecosystem:int64`, authWallet, appParam)
	api.HandleFunc("/appparams/{id}", AuthRequire(appParamsHandler)).Methods("GET")             // get(`appparams/:appid`, `?ecosystem:int64,?names:string`, authWallet, appParams)
	api.HandleFunc("/balance/{wallet}", AuthRequire(balanceHandler)).Methods("GET")             // get(`balance/:wallet`, `?ecosystem:int64`, authWallet, balance)
	api.HandleFunc("/contract/{name}", AuthRequire(contractInfoHandler)).Methods("GET")         // get(`contract/:name`, ``, authWallet, getContract)
	api.HandleFunc("/contracts", AuthRequire(contractsHandler)).Methods("GET")                  // get(`contracts`, `?limit ?offset:int64`, authWallet, getContracts)
	api.HandleFunc("/ecosystemparam/{name}", AuthRequire(ecosystemParamHandler)).Methods("GET") // get(`ecosystemparam/:name`, `?ecosystem:int64`, authWallet, ecosystemParam)
	api.HandleFunc("/ecosystemparams", AuthRequire(ecosystemParamsHandler)).Methods("GET")      // get(`ecosystemparams`, `?ecosystem:int64,?names:string`, authWallet, ecosystemParams)
	api.HandleFunc("/ecosystems", AuthRequire(ecosystemsHandler)).Methods("GET")                // get(`ecosystems`, ``, authWallet, ecosystems)
	api.HandleFunc("/getuid", uidHandler).Methods("GET")
	api.HandleFunc("/list/{name}", AuthRequire(listHandler)).Methods("GET")                           // get(`list/:name`, `?limit ?offset:int64,?columns:string`, authWallet, list)
	api.HandleFunc("/row/{name}/{id}", AuthRequire(rowHandler)).Methods("GET")                        // get(`row/:name/:id`, `?columns:string`, authWallet, row)
	api.HandleFunc("/interface/page/{name}", AuthRequire(pageRowHandler())).Methods("GET")            // get(`interface/page/:name`, ``, authWallet, getPageRow)
	api.HandleFunc("/interface/menu/{name}", AuthRequire(menuRowHandler())).Methods("GET")            // get(`interface/menu/:name`, ``, authWallet, getMenuRow)
	api.HandleFunc("/interface/block/{name}", AuthRequire(blockInterfaceRowHandler())).Methods("GET") // get(`interface/block/:name`, ``, authWallet, getBlockInterfaceRow)
	api.HandleFunc("/systemparams", AuthRequire(systemParamsHandler)).Methods("GET")                  // get(`systemparams`, `?names:string`, authWallet, systemParams)
	api.HandleFunc("/table/{name}", AuthRequire(tableHandler)).Methods("GET")                         // get(`table/:name`, ``, authWallet, table)
	api.HandleFunc("/tables", AuthRequire(tablesHandler)).Methods("GET")                              // get(`tables`, `?limit ?offset:int64`, authWallet, tables)
	api.HandleFunc("/txstatus/{hash}", AuthRequire(txstatusHandler)).Methods("GET")                   // get(`txstatus/:hash`, ``, authWallet, txstatus)
	api.HandleFunc("/test/{name}", testHandler).Methods("GET")                                        // get(`test/:name`, ``, getTest)
	api.HandleFunc("/history/{table}/{id}", AuthRequire(historyHandler)).Methods("GET")               // get(`history/:table/:id`, ``, authWallet, getHistory)
	api.HandleFunc("/block/{id}", blockInfoHandler).Methods("GET")
	api.HandleFunc("/maxblockid", maxBlockHandler).Methods("GET")
	api.HandleFunc("/version", versionHandler).Methods("GET")
	api.HandleFunc("/avatar/{ecosystem}/{member}", avatarHandler).Methods("GET")            // get(`avatar/:ecosystem/:member`, ``, getAvatar)
	api.HandleFunc("/config/{option}", configOptionHandler).Methods("GET")                  // get(`config/:option`, ``, getConfigOption)
	api.HandleFunc("/ecosystemname", ecosystemNameHandler).Methods("GET")                   //get("ecosystemname", "?id:int64", getEcosystemName)
	api.HandleFunc("/content/source/{name}", AuthRequire(getSourceHandler)).Methods("POST") // post(`content/source/:name`, ``, authWallet, getSource)
	api.HandleFunc("/content/page/{name}", AuthRequire(getPageHandler)).Methods("POST")     // post(`content/page/:name`, `?lang:string`, authWallet, getPage)
	api.HandleFunc("/content/menu/{name}", AuthRequire(getMenuHandler)).Methods("POST")     // post(`content/menu/:name`, `?lang:string`, authWallet, getMenu)
	api.HandleFunc("/content/hash/{name}", getPageHashHandler).Methods("POST")              // post(`content/hash/:name`, ``, getPageHash)
	// post(`vde/create`, ``, authWallet, vdeCreate)
	api.HandleFunc("/login", loginHandler).Methods("POST")                                                               // post(`login`, `?pubkey signature:hex,?key_id ?mobile:string,?ecosystem ?expire ?role_id:int64`, login)
	api.HandleFunc("/prepare/{name}", AuthRequire(contractHandlers.PrepareHandler)).Methods("POST")                      // post(`prepare/:name`, `?token_ecosystem:int64,?max_sum ?payover:string`, authWallet, contractHandlers.prepareContract)
	api.HandleFunc("/prepareMultiple", AuthRequire(contractHandlers.PrepareMultiHandler)).Methods("POST")                //post(`prepareMultiple`, `data:string`, authWallet, contractHandlers.prepareMultipleContract)
	api.HandleFunc("/txstatusMultiple", AuthRequire(txstatusMultiHandler)).Methods("POST")                               // post(`txstatusMultiple`, `data:string`, authWallet, txstatusMulti)
	api.HandleFunc("/contract/{request_id}", AuthRequire(contractHandlers.ContractHandler)).Methods("POST")              // post(`contract/:request_id`, `?pubkey signature:hex, time:string, ?token_ecosystem:int64,?max_sum ?payover:string`, authWallet, blockchainUpdatingState, contractHandlers.contract)
	api.HandleFunc("/contractMultiple/{request_id}", AuthRequire(contractHandlers.ContractMultiHandler)).Methods("POST") // post(`contractMultiple/:request_id`, `data:string`, authWallet, blockchainUpdatingState, contractHandlers.contractMulti)
	api.HandleFunc("/refresh", refreshHandler).Methods("POST")                                                           // post(`refresh`, `token:string,?expire:int64`, refresh)
	// post(`test/:name`, ``, getTest)
	api.HandleFunc("/content", jsonContentHandler).Methods("POST")              // post(`content`, `template ?source:string`, jsonContent)
	api.HandleFunc("/updnotificator", updateNotificatorHandler).Methods("POST") // post(`updnotificator`, `ids:string`, updateNotificator)

	// methodRoute(route, `POST`, `node/:name`, `?token_ecosystem:int64,?max_sum ?payover:string`, contractHandlers.nodeContract)
}
