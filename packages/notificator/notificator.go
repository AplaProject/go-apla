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

// Recipient is recipient struct
type Recipient struct {
	ID          int64
	EcosystemID int64
	IsVDE       bool
}

func (r *Recipient) ecosystemPrefix() string {
	id := strconv.FormatInt(r.EcosystemID, 10)
	if r.IsVDE {
		return id + "_vde"
	}
	return id
}

// NotificationStats storing notification stats data
type NotificationStats struct {
	recipients  sync.Map
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
		ecosystemPrefix := key.(string)
		ecosystemStats := value.(NotificationStats)

		notifs := getEcosystemNotifications(ecosystemPrefix, *ecosystemStats.lastNotifID, ecosystemStats)
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

			lni, ok := notifications.Load(ecosystemPrefix)
			ln := lni.(NotificationStats)
			if ok && *ln.lastNotifID < id {
				*ln.lastNotifID = id
				notifications.Store(ecosystemPrefix, ln)
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

func getEcosystemNotifications(ecosystemPrefix string, lastNotificationID int64, ns NotificationStats) []map[string]string {
	users := make([]int64, 0)
	ns.recipients.Range(func(key, value interface{}) bool {
		users = append(users, key.(int64))
		return true
	})

	rows, err := model.GetAllNotifocationsForEcosystem(ecosystemPrefix, lastNotificationID, users)
	if err != nil || len(rows) == 0 {
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all notifications")
		}
		return nil
	}
	return rows
}

// AddRecipient is subscribing user to notifications
func AddRecipient(r Recipient) {
	key := r.ecosystemPrefix()

	var ns NotificationStats
	ins, ok := notifications.Load(key)

	if !ok {
		ns = NotificationStats{
			recipients:  sync.Map{},
			lastNotifID: new(int64),
		}
	} else {
		ns = ins.(NotificationStats)
	}

	ns.recipients.Store(r.ID, 0)
	notifications.Store(key, ns)
}
