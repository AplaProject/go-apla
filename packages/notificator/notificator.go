package notificator

import (
	"encoding/json"
	"strconv"
	"sync"

	"github.com/AplaProject/go-apla/packages/converter"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/publisher"
	log "github.com/sirupsen/logrus"
)

type Ecosystem struct {
	EcosystemID int64
	IsVDE       bool
}

func (e *Ecosystem) prefix() string {
	prefix := strconv.FormatInt(e.EcosystemID, 10)
	if e.IsVDE {
		return prefix + "_vde"
	}
	return prefix
}

type notificationRecord struct {
	EcosystemID  int64 `json:"ecosystem"`
	IsVDE        bool  `json:"is_vde"`
	RoleID       int64 `json:"role_id"`
	RecordsCount int64 `json:"count"`
}

type lastMessagesKey struct {
	system Ecosystem
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

func (lm *lastMessages) get(system Ecosystem, user int64) ([]notificationRecord, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	res, ok := lm.stats[lastMessagesKey{system: system, user: user}]
	return res, ok
}

func (lm *lastMessages) set(system Ecosystem, user int64, newStats []notificationRecord) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.stats[lastMessagesKey{system: system, user: user}] = newStats
}

var (
	systemUsers       map[Ecosystem]*[]int64
	mu                sync.Mutex
	lastMessagesStats *lastMessages
)

func init() {
	systemUsers = make(map[Ecosystem]*[]int64)
	lastMessagesStats = newLastMessages()
}

func addUser(userID int64, ecosystem Ecosystem) {
	mu.Lock()
	defer mu.Unlock()

	val, ok := systemUsers[ecosystem]
	if ok {
		*val = append(*val, userID)
		return
	}

	val = &[]int64{userID}
	systemUsers[ecosystem] = val
}

// AddUser add user to send notifications
func AddUser(userID, systemID int64, isVDE bool) {
	ecosystem := Ecosystem{EcosystemID: systemID, IsVDE: isVDE}
	addUser(userID, ecosystem)
	UpdateNotifications(ecosystem, []int64{userID})
}

// UpdateNotifications send stats about unreaded messages to centrifugo for ecosystem
func UpdateNotifications(ecosystem Ecosystem, users []int64) {
	result, err := model.GetNotificationsCount(ecosystem.prefix(), users)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting notification count")
		return
	}

	notificationsStats := parseRecipientNotification(result, ecosystem)

	for recipient, stats := range notificationsStats {

		if oldStats, ok := lastMessagesStats.get(ecosystem, recipient); ok {
			if !statsChanged(oldStats, stats) {
				continue
			}
		}

		lastMessagesStats.set(ecosystem, recipient, *stats)

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
	for ecosystem, users := range systemUsers {
		UpdateNotifications(ecosystem, *users)
	}
}

func parseRecipientNotification(rows []map[string]string, ecosystem Ecosystem) map[int64]*[]notificationRecord {
	recipientNotifications := make(map[int64]*[]notificationRecord)

	for _, r := range rows {
		recipientID := converter.StrToInt64(r["recipient_id"])
		roleID := converter.StrToInt64(r["role_id"])
		count := converter.StrToInt64(r["cnt"])

		roleNotifications := notificationRecord{
			EcosystemID:  ecosystem.EcosystemID,
			IsVDE:        ecosystem.IsVDE,
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
