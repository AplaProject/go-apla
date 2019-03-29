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

package block

import (
	"sync"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
)

type banKey struct {
	Time time.Time   // banned till
	Bad  []time.Time // time of bad tx
}

var (
	banList = make(map[int64]banKey)
	mutex   = &sync.RWMutex{}
)

// IsBanned returns true if the key has been banned
func IsKeyBanned(keyID int64) bool {
	mutex.RLock()
	if ban, ok := banList[keyID]; ok {
		mutex.RUnlock()
		now := time.Now()
		if now.Before(ban.Time) {
			return true
		}
		for i := 0; i < conf.Config.BanKey.BadTx; i++ {
			if ban.Bad[i].Add(time.Duration(conf.Config.BanKey.BadTime) * time.Minute).After(now) {
				return false
			}
		}
		// Delete if time of all bad tx is old
		mutex.Lock()
		delete(banList, keyID)
		mutex.Unlock()
	} else {
		mutex.RUnlock()
	}
	return false
}

// BannedTill returns the time that the user has been banned till
func BannedTill(keyID int64) string {
	mutex.RLock()
	defer mutex.RUnlock()
	if ban, ok := banList[keyID]; ok {
		return ban.Time.Format(`2006-01-02 15:04:05`)
	}
	return ``
}

// BadTxForBan adds info about bad tx of the key
func BadTxForBan(keyID int64) {
	var (
		ban banKey
		ok  bool
	)
	mutex.Lock()
	defer mutex.Unlock()
	now := time.Now()
	if ban, ok = banList[keyID]; ok {
		var bMin, count int
		for i := 0; i < conf.Config.BanKey.BadTx; i++ {
			if ban.Bad[i].Add(time.Duration(conf.Config.BanKey.BadTime) * time.Minute).After(now) {
				count++
			}
			if i > bMin && ban.Bad[i].Before(ban.Bad[bMin]) {
				bMin = i
			}
		}
		ban.Bad[bMin] = now
		if count >= conf.Config.BanKey.BadTx-1 {
			ban.Time = now.Add(time.Duration(conf.Config.BanKey.BanTime) * time.Minute)
		}
	} else {
		ban = banKey{Bad: make([]time.Time, conf.Config.BanKey.BadTx)}
		ban.Bad[0] = time.Now()
	}
	banList[keyID] = ban
}
