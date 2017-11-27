package notificator

import (
	"encoding/json"
	"fmt"
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
type NotificationID int64

// NotificationStats storing notification stats data
type NotificationStats struct {
	UserIDs     map[UserID]NotificationID
	lastNotifID *NotificationID
}

type notificationRecord struct {
	RecipientID  UserID
	MaxNotifID   NotificationID
	RecordsCount int64
}

type Notifications struct {
	sync.Map
}

func AddUser(userID int64, ecosystemID int64) {
	if _, ok := notifications[EcosystemID(ecosystemID)]; !ok {
		notifications[EcosystemID(ecosystemID)] = NotificationStats{UserIDs: make(map[UserID]NotificationID), lastNotifID: new(NotificationID)}
	}
	notifications[EcosystemID(ecosystemID)].UserIDs[UserID(userID)] = 0
}

func SendNotifications() {
	for ecosystemID, ecosystemStats := range notifications {
		notifs := mapToStruct(getEcosystemNotifications(ecosystemID, ecosystemStats))
		for _, notif := range notifs {
			if notifications[ecosystemID].UserIDs[notif.RecipientID] >= notif.MaxNotifID {
				continue
			}
			ok, err := publisher.Write(int64(notif.RecipientID), notif.String())

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
			notifications[ecosystemID].UserIDs[notif.RecipientID] = notif.MaxNotifID
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

func getEcosystemNotifications(ecosystemID EcosystemID, userIDs NotificationStats) []map[string]string {
	users := make([]int64, 0)
	for userID := range userIDs.UserIDs {
		users = append(users, int64(userID))
	}
	rows, err := model.GetNotificationsCount(int64(ecosystemID), users)
	if err != nil || len(rows) == 0 {
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting notifications count")
		}
		return nil
	}
	return rows
}

func mapToStruct(dbData []map[string]string) []notificationRecord {
	var result []notificationRecord
	for _, record := range dbData {
		nf := &notificationRecord{}
		err := nf.ParseMap(record)
		if err != nil {
			continue
		}
		result = append(result, *nf)
	}
	return result
}

func (nr *notificationRecord) ParseMap(data map[string]string) error {
	convert := func(value string, errMessage string) (int64, error) {
		var result int64
		result, err := strconv.ParseInt(data[value], 10, 64)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err}).Error(errMessage)
		}
		return result, err
	}

	maxID, err := convert("max", "error converting max id")
	if err != nil {
		return err
	}

	count, err := convert("count", "error converting records count")
	if err != nil {
		return err
	}

	userID, err := convert("recipient_id", "error converting records count")
	if err != nil {
		return err
	}
	nr.MaxNotifID = NotificationID(maxID)
	nr.RecordsCount = count
	nr.RecipientID = UserID(userID)
	return nil
}

func (nr *notificationRecord) String() string {
	return fmt.Sprintf(`{"count": %d}`, nr.RecordsCount)
}
