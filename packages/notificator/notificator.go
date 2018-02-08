package notificator

import (
	"encoding/json"
	"strconv"

	"sync"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/publisher"
	log "github.com/sirupsen/logrus"
)

// EcosystemID is ecosystem id
type EcosystemID int64

// UserID is user id
type UserID int64

// NotificationStats storing notification stats data
type NotificationStats struct {
	userIDs     sync.Map
	lastNotifID *int64
}

type Notifications struct {
	sync.Map
}

//var notifications Notifications
var notifications Notifications

// SendNotifications is sending notifications
func SendNotifications() {
	notifications.Range(func(key, value interface{}) bool {
		ecosystemID := key.(EcosystemID)
		ecosystemStats := value.(NotificationStats)

		notifs := getEcosystemNotifications(ecosystemID, *ecosystemStats.lastNotifID, ecosystemStats)
		for _, notif := range notifs {
			userID, err := strconv.ParseInt(notif["recipient_id"], 10, 64)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConversionError, "value": notif["recipient_id"], "error": err}).Error("getting recipient_id")
				return false
			}

			data, err := mapToString(notif)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling notification")
				return false
			}

			ok, err := publisher.Write(userID, data)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing to centrifugo")
				return false
			}

			if !ok {
				log.WithFields(log.Fields{"type": consts.CentrifugoError, "error": err}).Error("writing to centrifugo")
				return false
			}

			id, err := strconv.ParseInt(notif["id"], 10, 64)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("conversion string to int64")
				return false
			}

			lni, ok := notifications.Load(ecosystemID)
			ln := lni.(NotificationStats)
			if ok && *ln.lastNotifID < id {
				*ln.lastNotifID = id
				notifications.Store(ecosystemID, ln)
			}
		}

		return true
	})
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
	userIDs.userIDs.Range(func(key, value interface{}) bool {
		users = append(users, int64(key.(UserID)))
		return true
	})

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
	eId := EcosystemID(ecosystemID)

	var ns NotificationStats
	ins, ok := notifications.Load(eId)

	if !ok {
		ns = NotificationStats{userIDs: sync.Map{}, lastNotifID: new(int64)}
	} else {
		ns = ins.(NotificationStats)
	}

	ns.userIDs.Store(UserID(userID), 0)
	notifications.Store(eId, ns)
}
