package notificator

import (
	"encoding/json"
	"sync"

	"github.com/AplaProject/go-apla/packages/converter"

	"github.com/AplaProject/go-apla/packages/consts"
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

	result, err := model.GetNotificationsCount(ecosystemID, users)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting notification count")
		return
	}

	notificationsStats := parseRecipientNotification(result, ecosystemID)

	for recipient, stats := range notificationsStats {

		if oldStats, ok := lastMessagesStats.get(ecosystemID, recipient); ok {
			if !statsChanged(oldStats, stats) {
				continue
			}
		}

		lastMessagesStats.set(ecosystemID, recipient, *stats)

		rawStats, err := json.Marshal(*stats)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("notification statistic")
			continue
		}

		ok, err := publisher.Write(recipient, string(rawStats))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing to centrifugo")
			continue
		}

		if !ok {
			log.WithFields(log.Fields{"type": consts.CentrifugoError, "error": err}).Error("writing to centrifugo")
		}
	}
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

func statsChanged(source []notificationRecord, new *[]notificationRecord) bool {

	var newRole bool

	for _, nRec := range *new {
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
