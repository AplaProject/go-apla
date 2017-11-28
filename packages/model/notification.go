package model

import (
	"strconv"
)

// GetAllNotifications is retrieving all notifications by params
func GetAllNotifications(ecosystemID int64, lastNotificationID int64, userIDs []int64) ([]map[string]string, error) {
	query := `select * from "` + strconv.FormatInt(ecosystemID, 10) +
		`_notifications" where closed = 0 and id > ` + strconv.FormatInt(lastNotificationID, 10) +
		` and recipient_id in (`
	for _, userID := range userIDs {
		query += strconv.FormatInt(int64(userID), 10) + ", "
	}
	query = query[:len(query)-2] + ");"
	return GetAll(query, -1)
}
