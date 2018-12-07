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

package notificator

import (
	"encoding/json"
	"sync"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/publisher"

	log "github.com/sirupsen/logrus"
)

type notificationRecord struct {
	EcosystemID  int64 `json:"ecosystem"`
	RoleID       int64 `json:"role_id"`
	RecordsCount int64 `json:"count"`
}

type lastMessagesKey struct {
	system int64
	user   int64
}

type lastMessages struct {
	mu    sync.RWMutex
	stats map[lastMessagesKey][]notificationRecord
}

func newLastMessages() *lastMessages {
	return &lastMessages{
		stats: map[lastMessagesKey][]notificationRecord{},
	}
}

func (lm *lastMessages) get(system, user int64) ([]notificationRecord, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	res, ok := lm.stats[lastMessagesKey{system: system, user: user}]
	return res, ok
}

func (lm *lastMessages) set(system, user int64, newStats []notificationRecord) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.stats[lastMessagesKey{system: system, user: user}] = newStats
}

func (lm *lastMessages) delete(system, user int64) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	delete(lm.stats, lastMessagesKey{system: system, user: user})
}

var (
	systemUsers       map[int64]*[]int64
	mu                sync.Mutex
	lastMessagesStats *lastMessages
)

func init() {
	systemUsers = make(map[int64]*[]int64)
	lastMessagesStats = newLastMessages()
}

// AddUser add user to send notifications
func AddUser(userID, systemID int64) {
	mu.Lock()
	defer mu.Unlock()

	val, ok := systemUsers[systemID]
	if ok {
		*val = append(*val, userID)
		return
	}

	val = &[]int64{userID}
	systemUsers[systemID] = val
}

// UpdateNotifications send stats about unreaded messages to centrifugo for ecosystem
func UpdateNotifications(ecosystemID int64, users []int64) {

	notificationsStats, err := getEcosystemNotificationStats(ecosystemID, users)
	if err != nil {
		return
	}

	for _, user := range users {
		oldStats, _ := lastMessagesStats.get(ecosystemID, user)
		var newStats []notificationRecord

		if ns, _ := notificationsStats[user]; ns != nil {
			newStats = *ns
		} else {
			newStats = nil
		}

		if !statsChanged(oldStats, newStats) {
			continue
		}

		if len(newStats) == 0 {
			for i := range oldStats {
				oldStats[i].RecordsCount = 0
			}

			lastMessagesStats.delete(ecosystemID, user)
			sendUserStats(user, oldStats)
			continue
		}

		lastMessagesStats.set(ecosystemID, user, newStats)
		sendUserStats(user, newStats)
	}
}

// UpdateRolesNotifications send stats about unreaded messages to centrifugo for ecosystem
func UpdateRolesNotifications(ecosystemID int64, roles []int64) {
	members, _ := model.GetRoleMembers(nil, ecosystemID, roles)
	UpdateNotifications(ecosystemID, members)
}

func getEcosystemNotificationStats(ecosystemID int64, users []int64) (map[int64]*[]notificationRecord, error) {
	result, err := model.GetNotificationsCount(ecosystemID, users)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting notification count")
		return nil, err
	}

	return parseRecipientNotification(result, ecosystemID), nil
}

// SendNotifications send stats about unreaded messages to centrifugo
func SendNotifications() {
	for ecosystemID, users := range systemUsers {
		UpdateNotifications(ecosystemID, *users)
	}
}

func parseRecipientNotification(rows []map[string]string, systemID int64) map[int64]*[]notificationRecord {
	recipientNotifications := make(map[int64]*[]notificationRecord)

	for _, r := range rows {
		recipientID := converter.StrToInt64(r["recipient_id"])
		if recipientID == 0 {
			continue
		}

		roleID := converter.StrToInt64(r["role_id"])
		count := converter.StrToInt64(r["cnt"])

		roleNotifications := notificationRecord{
			EcosystemID:  systemID,
			RoleID:       roleID,
			RecordsCount: count,
		}

		nr, ok := recipientNotifications[recipientID]
		if ok {
			*nr = append(*nr, roleNotifications)
			continue
		}

		records := []notificationRecord{
			roleNotifications,
		}

		recipientNotifications[recipientID] = &records
	}

	return recipientNotifications
}

func statsChanged(source, new []notificationRecord) bool {

	if len(source) != len(new) {
		return true
	}

	if len(new) == 0 {
		return false
	}

	var newRole bool

	for _, nRec := range new {
		newRole = true

		for _, sRec := range source {
			if sRec.RoleID == nRec.RoleID {
				newRole = false

				if sRec.RecordsCount != nRec.RecordsCount {
					return true
				}
			}
		}

		if newRole {
			return true
		}
	}
	return false
}

func sendUserStats(user int64, stats []notificationRecord) {
	rawStats, err := json.Marshal(stats)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("notification statistic")
	}

	ok, err := publisher.Write(user, string(rawStats))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing to centrifugo")
	}

	if !ok {
		log.WithFields(log.Fields{"type": consts.CentrifugoError, "error": err}).Error("writing to centrifugo")
	}
}

// SendNotificationsByRequest send stats by systemUsers one time
func SendNotificationsByRequest(systemUsers map[int64][]int64) {
	for ecosystemID, users := range systemUsers {
		stats, err := getEcosystemNotificationStats(ecosystemID, users)
		if err != nil {
			continue
		}

		for user, notifications := range stats {
			if notifications == nil {
				continue
			}

			sendUserStats(user, *notifications)
		}
	}
}
