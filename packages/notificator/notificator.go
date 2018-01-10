package notificator

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/publisher"
	log "github.com/sirupsen/logrus"
)

type notificationRecord struct {
	RoleID       int64 `json:"role_id"`
	RecordsCount int64 `json:"count"`
}

func (nr notificationRecord) String() string {
	return fmt.Sprintf(`{"role_id": %d, "count": %d}`, nr.RoleID, nr.RecordsCount)
}

// SendNotifications send stats about unreaded messages to centrifugo
func SendNotifications() {
	ecosystems, err := getEcosystemIDList()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting id list of ecosystems")
		return
	}

	for _, systemId := range ecosystems {
		result, err := model.GetNotificationsCount(systemId, nil)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting notification count")
			continue
		}

		notificationsStats, err := parseRecipientNotification(result)
		if err != nil {
			// error logged in parseRecipientNotification()
			continue
		}

		for recipient, stats := range notificationsStats {
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
				continue
			}
		}
	}
}

func getEcosystemIDList() ([]int64, error) {
	var idlist []int64

	db := model.GetDB(nil)
	rows, err := db.Raw("SELECT id FROM system_states").Rows()
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int64
		rows.Scan(&id)
		idlist = append(idlist, id)
	}

	return idlist, err
}

func parseRecipientNotification(rows []map[string]string) (map[int64]*[]notificationRecord, error) {
	recipientNotifications := make(map[int64]*[]notificationRecord)

	convert := func(dataRow map[string]string, value string, errMessage string) (int64, error) {
		var result int64
		result, err := strconv.ParseInt(dataRow[value], 10, 64)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error(errMessage)
		}
		return result, err
	}

	for _, r := range rows {
		recipientId, err := convert(r, "recipient_id", "error converting records count")
		if err != nil {
			return recipientNotifications, err
		}

		roleId, err := convert(r, "role_id", "error converting records count")
		if err != nil {
			return recipientNotifications, err
		}

		count, err := convert(r, "cnt", "error converting records count")
		if err != nil {
			return recipientNotifications, err
		}

		roleNotifications := notificationRecord{
			RoleID:       roleId,
			RecordsCount: count,
		}

		nr, ok := recipientNotifications[recipientId]
		if ok {
			*nr = append(*nr, roleNotifications)
			continue
		}

		records := []notificationRecord{
			roleNotifications,
		}

		recipientNotifications[recipientId] = &records
	}

	return recipientNotifications, nil
}
