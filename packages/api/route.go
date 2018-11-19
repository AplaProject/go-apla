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

	"github.com/GenesisKernel/go-genesis/packages/conf"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Route sets routing pathes
func setRoutes(r *mux.Router) {
	r.StrictSlash(true)
	r.Use(loggerMiddleware, recoverMiddleware, statsdMiddleware)

	// api router with prefix path
	api := r.PathPrefix("/api/v2").Subrouter()
	api.Use(nodeStateMiddleware, tokenMiddleware, clientMiddleware)

	api.HandleFunc("/data/{table}/{id}/{column}/{hash}", getDataHandler).Methods("GET")
	api.HandleFunc("/data/{prefix}_binaries/{id}/data/{hash}", getBinaryHandler).Methods("GET")
	api.HandleFunc("/avatar/{ecosystem}/{member}", getAvatarHandler).Methods("GET")

	api.HandleFunc("/contract/{name}", authRequire(getContractInfoHandler)).Methods("GET")
	api.HandleFunc("/contracts", authRequire(getContractsHandler)).Methods("GET")
	api.HandleFunc("/getuid", getUIDHandler).Methods("GET")
	api.HandleFunc("/keyinfo/{wallet}", getKeyInfoHandler).Methods("GET")
	api.HandleFunc("/list/{name}", authRequire(getListHandler)).Methods("GET")
	api.HandleFunc("/sections", authRequire(getSectionsHandler)).Methods("GET")
	api.HandleFunc("/row/{name}/{id}", authRequire(getRowHandler)).Methods("GET")
	api.HandleFunc("/interface/page/{name}", authRequire(getPageRowHandler)).Methods("GET")
	api.HandleFunc("/interface/menu/{name}", authRequire(getMenuRowHandler)).Methods("GET")
	api.HandleFunc("/interface/block/{name}", authRequire(getBlockInterfaceRowHandler)).Methods("GET")
	api.HandleFunc("/table/{name}", authRequire(getTableHandler)).Methods("GET")
	api.HandleFunc("/tables", authRequire(getTablesHandler)).Methods("GET")
	api.HandleFunc("/test/{name}", getTestHandler).Methods("GET", "POST")
	api.HandleFunc("/version", getVersionHandler).Methods("GET")
	api.HandleFunc("/config/{option}", getConfigOptionHandler).Methods("GET")

	api.HandleFunc("/content/source/{name}", authRequire(getSourceHandler)).Methods("POST")
	api.HandleFunc("/content/page/{name}", authRequire(getPageHandler)).Methods("POST")
	api.HandleFunc("/content/hash/{name}", authRequire(getPageHashHandler)).Methods("POST")
	api.HandleFunc("/content/menu/{name}", authRequire(getMenuHandler)).Methods("POST")
	api.HandleFunc("/content", authRequire(jsonContentHandler)).Methods("POST")
	api.HandleFunc("/login", loginHandler).Methods("POST")
	api.HandleFunc("/sendTx", authRequire(sendTxHandler)).Methods("POST")
	api.HandleFunc("/updnotificator", updateNotificatorHandler).Methods("POST")
	api.HandleFunc("/node/{name}", nodeContractHandler).Methods("POST")

	if !conf.Config.IsSupportingVDE() {
		api.HandleFunc("/txinfo/{hash}", authRequire(getTxInfoHandler)).Methods("GET")
		api.HandleFunc("/txinfomultiple", authRequire(getTxInfoMultiHandler)).Methods("GET")
		api.HandleFunc("/appparam/{appID}/{name}", authRequire(getAppParamHandler)).Methods("GET")
		api.HandleFunc("/appparams/{appID}", authRequire(getAppParamsHandler)).Methods("GET")
		api.HandleFunc("/history/{name}/{id}", authRequire(getHistoryHandler)).Methods("GET")
		api.HandleFunc("/balance/{wallet}", authRequire(getBalanceHandler)).Methods("GET")
		api.HandleFunc("/block/{id}", getBlockInfoHandler).Methods("GET")
		api.HandleFunc("/maxblockid", getMaxBlockHandler).Methods("GET")
		api.HandleFunc("/blocks", getBlocksTxInfoHandler).Methods("GET")
		api.HandleFunc("/detailed_blocks", getBlocksDetailedInfoHandler).Methods("GET")
		api.HandleFunc("/ecosystemparams", authRequire(getEcosystemParamsHandler)).Methods("GET")
		api.HandleFunc("/systemparams", authRequire(getSystemParamsHandler)).Methods("GET")
		api.HandleFunc("/ecosystems", authRequire(getEcosystemsHandler)).Methods("GET")
		api.HandleFunc("/ecosystemparam/{name}", authRequire(getEcosystemParamHandler)).Methods("GET")
		api.HandleFunc("/ecosystemname", getEcosystemNameHandler).Methods("GET")
		api.HandleFunc("/txstatus", authRequire(getTxStatusHandler)).Methods("POST")
	}
}

func NewRouter() http.Handler {
	r := mux.NewRouter()
	setRoutes(r)
	return r
}

func WithCors(h http.Handler) http.Handler {
	return handlers.CORS()(h)
}
