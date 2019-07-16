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

package daemons

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/AplaProject/go-apla/packages/api"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

const (
	errExternalNone    = iota // 0 - no error
	errExternalTx             // 1 - tx error
	errExternalAttempt        // 2 - attempt error
	errExternalTimeout        // 3 - timeout of getting txstatus

	maxAttempts           = 10
	statusTimeout         = 60
	externalDeamonTimeout = 2
	apiExt                = `/api/v2/`
)

var (
	nodePrivateKey []byte
	nodeKeyID      int64
	nodePublicKey  string
	authNet        = map[string]string{}
)

var enOnRun uint32

func loginNetwork(urlPath string) (connect *api.Connect, err error) {
	if len(nodePrivateKey) == 0 {
		var pubKey []byte
		if nodePrivateKey, err = utils.GetNodePrivateKey(); err != nil {
			return
		}
		if pubKey, err = crypto.PrivateToPublic(nodePrivateKey); err != nil {
			return
		}
		nodeKeyID = crypto.Address(pubKey)
		nodePublicKey = crypto.PubToHex(pubKey)
	}
	connect = &api.Connect{
		Auth:       authNet[urlPath],
		PrivateKey: nodePrivateKey,
		PublicKey:  nodePublicKey,
		Root:       urlPath,
	}
	if err = connect.Login(); err != nil {
		authNet[urlPath] = connect.Auth
	}
	return
}

func SendExternalTransaction() error {
	var (
		err     error
		connect *api.Connect
		delList []int64
		hash    string
	)

	toWait := map[string][]model.ExternalBlockchain{}
	incAttempt := func(id int64) {
		if err = model.IncExternalAttempt(id); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("IncAttempt")
		}
	}
	sendResult := func(item model.ExternalBlockchain, block, errCode int64, resText string) {
		defer func() {
			delList = append(delList, item.Id)
		}()
		if len(item.ResultContract) == 0 {
			return
		}
		if err := transaction.CreateContract(item.ResultContract, nodeKeyID,
			map[string]interface{}{
				"Status": errCode,
				"Msg":    resText,
				"Block":  block,
				"UID":    item.Uid,
			}, nodePrivateKey); err != nil {
			log.WithFields(log.Fields{"type": consts.ContractError, "err": err}).Error("CreateContract")
		}
	}
	list, err := model.GetExternalList()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("GetExternalList")
		return err
	}
	timeOut := time.Now().Unix() - 10*(syspar.GetGapsBetweenBlocks()+
		syspar.GetMaxBlockGenerationTime()/1000)
	for _, item := range list {
		root := item.Url + apiExt

		if item.Sent == 0 {
			if timeOut > item.TxTime {
				delList = append(delList, item.Id)
				continue
			}
			if connect, err = loginNetwork(root); err != nil {
				log.WithFields(log.Fields{"type": consts.AccessDenied, "error": err}).Error("loginNetwork")
				return err
			}
			values := url.Values{"UID": {item.Uid}}

			var params map[string]interface{}
			if err = json.Unmarshal([]byte(item.Value), &params); err != nil {
				log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Unmarshal params")
				delList = append(delList, item.Id)
				continue
			}
			for key, val := range params {
				values[key] = []string{fmt.Sprint(val)}
			}
			values["nowait"] = []string{"1"}
			values["txtime"] = []string{converter.Int64ToStr(item.TxTime)}
			_, hash, err = connect.PostTxResult(item.ExternalContract, &values)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("PostContract")
				if item.Attempts >= maxAttempts-1 {
					sendResult(item, 0, errExternalAttempt, ``)
				} else {
					incAttempt(item.Id)
				}
			} else {
				bHash, err := hex.DecodeString(hash)
				if err != nil {
					log.WithFields(log.Fields{"type": consts.ParseError, "error": err}).Error("DecodeHex")
					incAttempt(item.Id)
				} else {
					model.HashExternalTx(item.Id, bHash)
				}
			}
		} else {
			toWait[item.Url] = append(toWait[item.Url], item)
		}
	}
	for _, waitList := range toWait {
		if connect, err = loginNetwork(waitList[0].Url + apiExt); err != nil {
			log.WithFields(log.Fields{"type": consts.AccessDenied, "error": err}).Error("loginNetwork")
			continue
		}
		var hashes []string
		for _, item := range waitList {
			hashes = append(hashes, hex.EncodeToString(item.Hash))
		}
		results, err := connect.WaitTxList(hashes)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("WaitTxList")
			continue
		}
		timeOut = time.Now().Unix() - statusTimeout
		for _, item := range waitList {
			if result, ok := results[hex.EncodeToString(item.Hash)]; ok {
				errCode := int64(errExternalNone)
				if result.BlockID == 0 {
					errCode = errExternalTx
				}
				sendResult(item, result.BlockID, errCode, result.Msg)
			} else if timeOut > item.TxTime {
				sendResult(item, 0, errExternalTimeout, ``)
			}
		}
	}
	if len(delList) > 0 {
		if err = model.DelExternalList(delList); err != nil {
			return err
		}
	}
	return nil
}

// ExternalNetwork sends txinfo to the external network
func ExternalNetwork(ctx context.Context, d *daemon) error {
	if !atomic.CompareAndSwapUint32(&enOnRun, 0, 1) {
		return nil
	}
	defer func() {
		atomic.StoreUint32(&enOnRun, 0)
	}()
	d.sleepTime = externalDeamonTimeout * time.Second
	return SendExternalTransaction()
}
