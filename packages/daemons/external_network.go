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
	"time"

	"github.com/AplaProject/go-apla/packages/api"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/converter"
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

func loginNetwork(netName, urlPath string) error {
	var (
		err  error
		sign []byte
	)
	if len(nodePrivateKey) == 0 {
		var pubKey []byte
		if nodePrivateKey, err = utils.GetNodePrivateKey(); err != nil {
			return err
		}
		if pubKey, err = crypto.PrivateToPublic(nodePrivateKey); err != nil {
			return err
		}
		nodePublicKey = crypto.PubToHex(pubKey)
	}

	var ret api.GetUIDResult
	err = api.SendGet(urlPath+`getuid`, authNet[netName], nil, &ret)
	if err != nil {
		return err
	}
	if len(ret.UID) == 0 {
		return nil
	}
	authNet[netName] = ret.Token
	sign, err = crypto.SignString(hex.EncodeToString(nodePrivateKey), `LOGIN`+ret.NetworkID+ret.UID)
	if err != nil {
		return err
	}
	form := url.Values{"pubkey": {nodePublicKey}, "signature": {hex.EncodeToString(sign)},
		`ecosystem`: {`1`}, "role_id": {"0"}}
	var logret api.LoginResult
	err = api.SendPost(urlPath+`login`, authNet[netName], &form, &logret)
	authNet[netName] = logret.Token
	return err
}

func SendToNetwork() error {
	var (
		external map[string]smart.ExternalNetInfo
		err      error
		ok       bool
		duration time.Duration
		prevTime time.Time
	)
	if err = json.Unmarshal([]byte(syspar.SysString(syspar.ExternalBlockchain)), &external); err != nil {
		return err
	}
	for key, netInfo := range external {
		if len(netInfo.Interval) < 2 {
			netInfo.Interval = `1h`
		}
		typeInterval := netInfo.Interval[len(netInfo.Interval)-1]
		interval := converter.StrToInt64(netInfo.Interval[:len(netInfo.Interval)-1])
		if interval == 0 {
			interval = 1
		}
		switch typeInterval {
		case 'm':
			duration = time.Duration(interval) * time.Minute
		case 's':
			duration = time.Duration(interval) * time.Second
		default:
			duration = time.Duration(interval) * time.Hour
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
		if err = loginNetwork(key, root); err != nil {
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
		id, _, err := api.PostTxResult(root, authNet[key], netInfo.Contract, nodePrivateKey,
			&url.Values{"List": {string(out)}})
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
	d.sleepTime = 30 * time.Second
	err := SendToNetwork()
	if err != nil {
		fmt.Println(`ERROR`, err)
	}
	return err
}
