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

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// setRoutes sets routing paths
func setRoutes(router *mux.Router) {
	router.StrictSlash(true)
	router.Use(LoggerMiddleware, RecoverMiddleware, StatsdMiddleware)

	// api router with prefix path
	api := router.PathPrefix("/api/v2").Subrouter()
	api.Use(NodeStateMiddleware, TokenMiddleware, ClientMiddleware)

	contractHandlers := &contractHandlers{
		requests: tx.NewRequestBuffer(consts.TxRequestExpire),
	}

	api.HandleFunc("/data/{table}/{id}/{column}/{hash}", dataHandler).Methods("GET")
	api.HandleFunc("/data/{prefix}_binaries/{id}/data/{hash}", binaryHandler).Methods("GET")
	api.HandleFunc("/contract/{name}", AuthRequire(contractInfoHandler)).Methods("GET")
	api.HandleFunc("/contracts", AuthRequire(contractsHandler)).Methods("GET")
	api.HandleFunc("/getuid", uidHandler).Methods("GET")
	api.HandleFunc("/list/{name}", AuthRequire(listHandler)).Methods("GET")
	api.HandleFunc("/row/{name}/{id}", AuthRequire(rowHandler)).Methods("GET")
	api.HandleFunc("/interface/page/{name}", AuthRequire(pageRowHandler())).Methods("GET")
	api.HandleFunc("/interface/menu/{name}", AuthRequire(menuRowHandler())).Methods("GET")
	api.HandleFunc("/interface/block/{name}", AuthRequire(blockInterfaceRowHandler())).Methods("GET")
	api.HandleFunc("/systemparams", AuthRequire(systemParamsHandler)).Methods("GET")
	api.HandleFunc("/table/{name}", AuthRequire(tableHandler)).Methods("GET")
	api.HandleFunc("/tables", AuthRequire(tablesHandler)).Methods("GET")
	api.HandleFunc("/test/{name}", testHandler).Methods("GET")
	api.HandleFunc("/version", versionHandler).Methods("GET")
	api.HandleFunc("/avatar/{ecosystem}/{member}", avatarHandler).Methods("GET")
	api.HandleFunc("/config/{option}", configOptionHandler).Methods("GET")
	api.HandleFunc("/ecosystemname", ecosystemNameHandler).Methods("GET")
	api.HandleFunc("/content/source/{name}", AuthRequire(getSourceHandler)).Methods("POST")
	api.HandleFunc("/content/page/{name}", AuthRequire(getPageHandler)).Methods("POST")
	api.HandleFunc("/content/menu/{name}", AuthRequire(getMenuHandler)).Methods("POST")
	api.HandleFunc("/content/hash/{name}", getPageHashHandler).Methods("POST")
	api.HandleFunc("/login", loginHandler).Methods("POST")
	api.HandleFunc("/prepare", AuthRequire(contractHandlers.PrepareHandler)).Methods("POST")
	api.HandleFunc("/contract/{request_id}", AuthRequire(contractHandlers.ContractMultiHandler)).Methods("POST")
	api.HandleFunc("/refresh", refreshHandler).Methods("POST")
	api.HandleFunc("/content", jsonContentHandler).Methods("POST")
	api.HandleFunc("/updnotificator", updateNotificatorHandler).Methods("POST")

	if isVDEMode() {
		api.HandleFunc("/node/{name}", AuthRequire(contractHandlers.ContractNodeHandler)).Methods("POST")
	} else {
		api.HandleFunc("/txstatus", AuthRequire(txstatusMultiHandler)).Methods("POST")
		api.HandleFunc("/appparam/{id}/{name}", AuthRequire(appParamHandler)).Methods("GET")
		api.HandleFunc("/appparams/{id}", AuthRequire(appParamsHandler)).Methods("GET")
		api.HandleFunc("/balance/{wallet}", AuthRequire(balanceHandler)).Methods("GET")
		api.HandleFunc("/history/{table}/{id}", AuthRequire(historyHandler)).Methods("GET")
		api.HandleFunc("/block/{id}", blockInfoHandler).Methods("GET")
		api.HandleFunc("/maxblockid", maxBlockHandler).Methods("GET")
		api.HandleFunc("/ecosystemparam/{name}", AuthRequire(ecosystemParamHandler)).Methods("GET")
		api.HandleFunc("/ecosystemparams", AuthRequire(ecosystemParamsHandler)).Methods("GET")
		api.HandleFunc("/ecosystems", AuthRequire(ecosystemsHandler)).Methods("GET")
	}
}

func NewRouter() *mux.Router {
	router := mux.NewRouter()
	setRoutes(router)
	return router
}

func UseCORS(router *mux.Router) http.Handler {
	if isVDEMode() {
		return router
	}
	return handlers.CORS()(router)
}
