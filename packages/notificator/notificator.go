package notificator

import (
	"encoding/json"
	"strconv"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/publisher"
	log "github.com/sirupsen/logrus"
)

type EcosystemID int64
type UserID int64

type NotificationStats struct {
	UserIDs     map[UserID]int64
	lastNotifID *int64
}

var notifications map[EcosystemID]NotificationStats

func init() {
	notifications = make(map[EcosystemID]NotificationStats)
}

func SendNotifications() {
	for ecosystemID, ecosystemStats := range notifications {
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
			if *notifications[ecosystemID].lastNotifID < id {
				*notifications[ecosystemID].lastNotifID = id
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

func AddUser(userID int64, ecosystemID int64) {
	if _, ok := notifications[EcosystemID(ecosystemID)]; !ok {
		notifications[EcosystemID(ecosystemID)] = NotificationStats{UserIDs: make(map[UserID]int64), lastNotifID: new(int64)}
	}
	notifications[EcosystemID(ecosystemID)].UserIDs[UserID(userID)] = 0
}
