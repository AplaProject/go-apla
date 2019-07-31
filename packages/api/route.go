// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package api

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const corsMaxAge = 600

type Router struct {
	main        *mux.Router
	apiVersions map[string]*mux.Router
}

func (r Router) GetAPI() *mux.Router {
	return r.main
}

func (r Router) GetAPIVersion(preffix string) *mux.Router {
	return r.apiVersions[preffix]
}

func (r Router) NewVersion(preffix string) *mux.Router {
	api := r.main.PathPrefix(preffix).Subrouter()
	r.apiVersions[preffix] = api
	return api
}

// Route sets routing pathes
func (m Mode) SetCommonRoutes(r Router) {
	api := r.NewVersion("/api/v2")

	api.Use(nodeStateMiddleware, tokenMiddleware, m.clientMiddleware)

	api.HandleFunc("/data/{prefix}_binaries/{id}/data/{hash}", getBinaryHandler).Methods("GET")
	api.HandleFunc("/data/{table}/{id}/{column}/{hash}", getDataHandler).Methods("GET")
	api.HandleFunc("/avatar/{ecosystem}/{account}", getAvatarHandler).Methods("GET")
	api.HandleFunc("/auth/status", getAuthStatus).Methods("GET")

	api.HandleFunc("/contract/{name}", authRequire(getContractInfoHandler)).Methods("GET")
	api.HandleFunc("/contracts", authRequire(getContractsHandler)).Methods("GET")
	api.HandleFunc("/getuid", getUIDHandler).Methods("GET")
	api.HandleFunc("/keyinfo/{wallet}", m.getKeyInfoHandler).Methods("GET")
	api.HandleFunc("/list/{name}", authRequire(getListHandler)).Methods("GET")
	api.HandleFunc("/network", getNetworkHandler).Methods("GET")
	api.HandleFunc("/sections", authRequire(getSectionsHandler)).Methods("GET")
	api.HandleFunc("/row/{name}/{id}", authRequire(getRowHandler)).Methods("GET")
	api.HandleFunc("/row/{name}/{column}/{id}", authRequire(getRowHandler)).Methods("GET")
	api.HandleFunc("/interface/page/{name}", authRequire(getPageRowHandler)).Methods("GET")
	api.HandleFunc("/interface/menu/{name}", authRequire(getMenuRowHandler)).Methods("GET")
	api.HandleFunc("/interface/block/{name}", authRequire(getBlockInterfaceRowHandler)).Methods("GET")
	api.HandleFunc("/table/{name}", authRequire(getTableHandler)).Methods("GET")
	api.HandleFunc("/tables", authRequire(getTablesHandler)).Methods("GET")
	api.HandleFunc("/test/{name}", getTestHandler).Methods("GET", "POST")
	api.HandleFunc("/version", getVersionHandler).Methods("GET")
	api.HandleFunc("/config/{option}", getConfigOptionHandler).Methods("GET")

	api.HandleFunc("/page/validators_count/{name}", getPageValidatorsCountHandler).Methods("GET")
	api.HandleFunc("/content/source/{name}", authRequire(getSourceHandler)).Methods("POST")
	api.HandleFunc("/content/page/{name}", authRequire(getPageHandler)).Methods("POST")
	api.HandleFunc("/content/hash/{name}", getPageHashHandler).Methods("POST")
	api.HandleFunc("/content/menu/{name}", authRequire(getMenuHandler)).Methods("POST")
	api.HandleFunc("/content", jsonContentHandler).Methods("POST")
	api.HandleFunc("/login", m.loginHandler).Methods("POST")
	api.HandleFunc("/sendTx", authRequire(m.sendTxHandler)).Methods("POST")
	api.HandleFunc("/node/{name}", nodeContractHandler).Methods("POST")
	api.HandleFunc("/txstatus", authRequire(getTxStatusHandler)).Methods("POST")
	api.HandleFunc("/metrics/blocks", blocksCountHandler).Methods("GET")
	api.HandleFunc("/metrics/transactions", txCountHandler).Methods("GET")
	api.HandleFunc("/metrics/ecosystems", m.ecosysCountHandler).Methods("GET")
	api.HandleFunc("/metrics/keys", keysCountHandler).Methods("GET")
	api.HandleFunc("/metrics/mem", memStatHandler).Methods("GET")
	api.HandleFunc("/metrics/ban", banStatHandler).Methods("GET")
}

func (m Mode) SetBlockchainRoutes(r Router) {
	api := r.GetAPIVersion("/api/v2")
	api.HandleFunc("/metrics/fullnodes", fullNodesCountHandler).Methods("GET")
	api.HandleFunc("/txinfo/{hash}", authRequire(getTxInfoHandler)).Methods("GET")
	api.HandleFunc("/txinfomultiple", authRequire(getTxInfoMultiHandler)).Methods("GET")
	api.HandleFunc("/appparam/{appID}/{name}", authRequire(m.GetAppParamHandler)).Methods("GET")
	api.HandleFunc("/appparams/{appID}", authRequire(m.getAppParamsHandler)).Methods("GET")
	api.HandleFunc("/appcontent/{appID}", authRequire(m.getAppContentHandler)).Methods("GET")
	api.HandleFunc("/history/{name}/{id}", authRequire(getHistoryHandler)).Methods("GET")
	api.HandleFunc("/balance/{wallet}", authRequire(m.getBalanceHandler)).Methods("GET")
	api.HandleFunc("/block/{id}", getBlockInfoHandler).Methods("GET")
	api.HandleFunc("/maxblockid", getMaxBlockHandler).Methods("GET")
	api.HandleFunc("/blocks", getBlocksTxInfoHandler).Methods("GET")
	api.HandleFunc("/detailed_blocks", getBlocksDetailedInfoHandler).Methods("GET")
	api.HandleFunc("/ecosystemparams", authRequire(m.getEcosystemParamsHandler)).Methods("GET")
	api.HandleFunc("/systemparams", authRequire(getSystemParamsHandler)).Methods("GET")
	api.HandleFunc("/ecosystemparam/{name}", authRequire(m.getEcosystemParamHandler)).Methods("GET")
	api.HandleFunc("/ecosystemname", getEcosystemNameHandler).Methods("GET")
}

func NewRouter(m Mode) Router {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.Use(loggerMiddleware, recoverMiddleware, statsdMiddleware)

	api := Router{
		main:        r,
		apiVersions: make(map[string]*mux.Router),
	}
	m.SetCommonRoutes(api)
	return api
}

func WithCors(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type", "X-Requested-With"}),
		handlers.MaxAge(corsMaxAge),
	)(h)
}
