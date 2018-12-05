// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package api

import (
	"net/http"

	"github.com/AplaProject/go-apla/packages/conf"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const corsMaxAge = 600

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
	api.HandleFunc("/content/hash/{name}", getPageHashHandler).Methods("POST")
	api.HandleFunc("/content/menu/{name}", authRequire(getMenuHandler)).Methods("POST")
	api.HandleFunc("/content", jsonContentHandler).Methods("POST")
	api.HandleFunc("/login", loginHandler).Methods("POST")
	api.HandleFunc("/sendTx", authRequire(sendTxHandler)).Methods("POST")
	api.HandleFunc("/updnotificator", updateNotificatorHandler).Methods("POST")
	api.HandleFunc("/node/{name}", nodeContractHandler).Methods("POST")

	if !conf.Config.IsSupportingVDE() {
		api.HandleFunc("/txinfo/{hash}", authRequire(getTxInfoHandler)).Methods("GET")
		api.HandleFunc("/txinfomultiple", authRequire(getTxInfoMultiHandler)).Methods("GET")
		api.HandleFunc("/appparam/{appID}/{name}", authRequire(getAppParamHandler)).Methods("GET")
		api.HandleFunc("/appparams/{appID}", authRequire(getAppParamsHandler)).Methods("GET")
		api.HandleFunc("/appcontent/{appID}", authRequire(getAppContentHandler)).Methods("GET")
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
	return handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type", "X-Requested-With"}),
		handlers.MaxAge(corsMaxAge),
	)(h)
}
