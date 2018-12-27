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

package publisher

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/centrifugal/gocent"
	log "github.com/sirupsen/logrus"
)

type ClientsChannels struct {
	storage map[int64]string
	sync.RWMutex
}

func (cn *ClientsChannels) Set(id int64, s string) {
	cn.Lock()
	defer cn.Unlock()
	cn.storage[id] = s
}

func (cn *ClientsChannels) Get(id int64) string {
	cn.RLock()
	defer cn.RUnlock()
	return cn.storage[id]
}

var (
	clientsChannels   = ClientsChannels{storage: make(map[int64]string)}
	centrifugoTimeout = time.Second * 5
	publisher         *gocent.Client
	config            conf.CentrifugoConfig
)

// InitCentrifugo client
func InitCentrifugo(config conf.CentrifugoConfig) {
	publisher = gocent.New(config.GocentConfig())
}

func GetHMACSign(userID int64) (string, string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	secret, err := crypto.GetHMACWithTimestamp(config.Secret, strconv.FormatInt(userID, 10), timestamp)

	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("HMAC getting error")
		return "", "", err
	}

	result := hex.EncodeToString(secret)
	clientsChannels.Set(userID, result)
	return result, timestamp, nil
}

// Write is publishing data to server
func Write(userID int64, data string) error {
	ctx, cancel := context.WithTimeout(context.Background(), centrifugoTimeout)
	defer cancel()
	return publisher.Publish(ctx, "client"+strconv.FormatInt(userID, 10), []byte(data))
}

// GetInfo returns Stats
func GetInfo() (gocent.InfoResult, error) {
	if publisher == nil {
		return gocent.InfoResult{}, fmt.Errorf("publisher not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), centrifugoTimeout)
	defer cancel()
	return publisher.Info(ctx)
}
