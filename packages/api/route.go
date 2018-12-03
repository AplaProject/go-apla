// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
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
//
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
	"strings"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"

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

	route.Handle(`OPTIONS`, consts.ApiPath+`*name`, optionsHandler())
	route.Handle(`GET`, consts.ApiPath+`data/:table/:id/:column/:hash`, dataHandler())

	get(`contract/:name`, ``, authWallet, getContract)
	get(`contracts`, `?limit ?offset:int64`, authWallet, getContracts)
	get(`getuid`, ``, getUID)
	get(`keyinfo/:wallet`, ``, keyInfo)
	get(`list/:name`, `?limit ?offset:int64,?columns:string`, authWallet, list)
	get(`sections`, `?limit ?offset:int64,?lang:string`, authWallet, sections)
	get(`row/:name/:id`, `?columns:string`, authWallet, row)
	get(`interface/page/:name`, ``, authWallet, getPageRow)
	get(`interface/menu/:name`, ``, authWallet, getMenuRow)
	get(`interface/block/:name`, ``, authWallet, getBlockInterfaceRow)
	// get(`systemparams`, `?names:string`, authWallet, systemParams)
	get(`table/:name`, ``, authWallet, table)
	get(`tables`, `?limit ?offset:int64`, authWallet, tables)
	get(`test/:name`, ``, getTest)
	get(`version`, ``, getVersion)
	get(`avatar/:ecosystem/:member`, ``, getAvatar)
	get(`config/:option`, ``, getConfigOption)
	post(`content/source/:name`, ``, authWallet, getSource)
	post(`content/page/:name`, `?lang:string`, authWallet, getPage)
	post(`content/menu/:name`, `?lang:string`, authWallet, getMenu)
	post(`content/hash/:name`, `?lang:string`, getPageHash)
	post(`login`, `?pubkey signature:hex,?key_id ?mobile:string,?ecosystem ?expire ?role_id:int64`, login)
	post(`sendTx`, ``, authWallet, sendTx)
	post(`test/:name`, ``, getTest)
	post(`content`, `template ?source:string`, jsonContent)
	post(`updnotificator`, `ids:string`, updateNotificator)
	methodRoute(route, `POST`, `node/:name`, `?token_ecosystem:int64,?max_sum ?payover:string`, nodeContract)

	if !conf.Config.IsSupportingVDE() {
		get(`txinfo/:hash`, `?contractinfo:int64`, authWallet, txinfo)
		get(`txinfomultiple`, `data:string,?contractinfo:int64`, authWallet, txinfoMulti)
		get(`appparam/:appid/:name`, `?ecosystem:int64`, authWallet, appParam)
		get(`appparams/:appid`, `?ecosystem:int64,?names:string`, authWallet, appParams)
		get(`history/:table/:id`, ``, authWallet, getHistory)
		get(`balance/:wallet`, `?ecosystem:int64`, authWallet, balance)
		get(`block/:id`, ``, getBlockInfo)
		get(`maxblockid`, ``, getMaxBlockID)
		get("blocks", "block_id ?count:int64", getBlocksTxInfo)
		get("detailed_blocks", "block_id ?count:int64", getBlocksDetailedInfo)
		get(`ecosystemparams`, `?ecosystem:int64,?names:string`, authWallet, ecosystemParams)
		get(`systemparams`, `?names:string`, authWallet, systemParams)
		get(`ecosystems`, ``, authWallet, ecosystems)
		get(`ecosystemparam/:name`, `?ecosystem:int64`, authWallet, ecosystemParam)
		get("ecosystemname", "?id:int64", getEcosystemName)
		post(`txstatus`, `data:string`, authWallet, txstatus)
	}
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
