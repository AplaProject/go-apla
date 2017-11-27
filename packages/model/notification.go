package model

import (
	"strconv"
)

func GetNotificationsCount(ecosystemID int64, userIDs []int64) ([]map[string]string, error) {
	query := `select count(recipient_id), max(id), recipient_id from "` + strconv.FormatInt(ecosystemID, 10) +
		`_notifications" where recipient_id in (`
	for _, userID := range userIDs {
		query += strconv.FormatInt(int64(userID), 10) + ", "
	}
	query = query[:len(query)-2] + `) group by "recipient_id";`

	return GetAll(query, -1)
}
