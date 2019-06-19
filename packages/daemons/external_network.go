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
	"encoding/json"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/AplaProject/go-apla/packages/api"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils"
)

const (
	countTx = 100 // maximum records for sending
)

var (
	nodePrivateKey []byte
	nodePublicKey  string
	timeNet        = map[string]time.Time{}
	authNet        = map[string]string{}
)

var enOnRun uint32

func loginNetwork(netName, urlPath string) (connect *api.Connect, err error) {
	if len(nodePrivateKey) == 0 {
		var pubKey []byte
		if nodePrivateKey, err = utils.GetNodePrivateKey(); err != nil {
			return
		}
		if pubKey, err = crypto.PrivateToPublic(nodePrivateKey); err != nil {
			return
		}
		nodePublicKey = crypto.PubToHex(pubKey)
	}
	connect = &api.Connect{
		Auth:       authNet[netName],
		PrivateKey: nodePrivateKey,
		PublicKey:  nodePublicKey,
		Root:       urlPath,
	}
	if err = connect.Login(); err != nil {
		authNet[netName] = connect.Auth
	}
	return
}

func SendToNetwork() error {
	var (
		external map[string]smart.ExternalNetInfo
		err      error
		ok       bool
		connect  *api.Connect
		duration time.Duration
		prevTime time.Time
	)
	if err = json.Unmarshal([]byte(syspar.SysString(syspar.ExternalBlockchain)), &external); err != nil {
		return err
	}
	for key, netInfo := range external {
		duration, err = time.ParseDuration(netInfo.Interval)
		if err != nil {
			continue
		}
		if prevTime, ok = timeNet[key]; ok {
			if time.Now().Before(prevTime.Add(duration)) {
				continue
			}
		}
		list, err := model.GetExternalList(key, countTx)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			continue
		}
		root := netInfo.URL + `/api/v2/`
		if connect, err = loginNetwork(key, root); err != nil {
			return err
		}
		outList := make([]interface{}, 0, len(list))
		sentList := make([]int64, 0, len(list))
		for _, item := range list {
			var mitem interface{}
			if err = json.Unmarshal([]byte(item.Value), &mitem); err != nil {
				continue
			}
			outList = append(outList, mitem)
			sentList = append(sentList, item.ID)
		}
		out, err := json.Marshal(outList)
		if err != nil {
			continue
		}
		id, _, err := connect.PostTxResult(netInfo.Contract, &url.Values{"List": {string(out)}})
		timeNet[key] = time.Now()
		if id != 0 && err == nil {
			if err = model.DelExternalList(sentList); err != nil {
				return err
			}
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
	d.sleepTime = 30 * time.Second
	return SendToNetwork()
}
