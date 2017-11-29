package notificator

import (
	"encoding/json"
	"strconv"

	"sync"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/publisher"
	log "github.com/sirupsen/logrus"
)

// EcosystemID is ecosystem id
type EcosystemID int64

// UserID is user id
type UserID int64

// NotificationStats storing notification stats data
type NotificationStats struct {
	UserIDs     map[UserID]int64
	lastNotifID *int64
}

type ConcurrentNotifications struct {
	storage map[EcosystemID]NotificationStats
	sync.Mutex
}

var notifications ConcurrentNotifications

func init() {
	notifications = ConcurrentNotifications{storage: make(map[EcosystemID]NotificationStats)}
}

// SendNotifications is sending notifications
func SendNotifications() {
	for ecosystemID, ecosystemStats := range notifications.storage {
		notifs := getEcosystemNotifications(ecosystemID, *ecosystemStats.lastNotifID, ecosystemStats)
		for _, notif := range notifs {
			userID, err := strconv.ParseInt(notif["recipient_id"], 10, 64)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConversionError, "value": notif["recipient_id"], "error": err}).Error("getting recipient_id")
				return
			}
			data, err := mapToString(notif)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling notification")
				return
			}
			ok, err := publisher.Write(userID, data)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing to centrifugo")
				return
			}

			if !ok {
				log.WithFields(log.Fields{"type": consts.CentrifugoError, "error": err}).Error("writing to centrifugo")
				return
			}
			id, _ := strconv.ParseInt(notif["id"], 10, 64)
			if *notifications.storage[ecosystemID].lastNotifID < id {
				notifications.Lock()
				*notifications.storage[ecosystemID].lastNotifID = id
				notifications.Unlock()
			}
		}
	}
}

func mapToString(value map[string]string) (string, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func getEcosystemNotifications(ecosystemID EcosystemID, lastNotificationID int64, userIDs NotificationStats) []map[string]string {
	users := make([]int64, 0)
	for userID := range userIDs.UserIDs {
		users = append(users, int64(userID))
	}
	rows, err := model.GetAllNotifications(int64(ecosystemID), lastNotificationID, users)
	if err != nil || len(rows) == 0 {
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all notifications")
		}
		return nil
	}
	return rows
}

// AddUser is subscribing user to notifications
func AddUser(userID int64, ecosystemID int64) {
	if _, ok := notifications.storage[EcosystemID(ecosystemID)]; !ok {
		notifications.Lock()
		notifications.storage[EcosystemID(ecosystemID)] = NotificationStats{UserIDs: make(map[UserID]int64), lastNotifID: new(int64)}
		notifications.Unlock()
	}
	notifications.storage[EcosystemID(ecosystemID)].UserIDs[UserID(userID)] = 0
}
