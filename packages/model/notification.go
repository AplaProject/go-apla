package model

import "github.com/AplaProject/go-apla/packages/converter"

// GetAllNotificationsForEcosystem is retrieving all notifications by params
func GetAllNotificationsForEcosystem(prefix string, lastNotificationID int64, userIDs []int64) ([]map[string]string, error) {
	tableName := converter.EscapeName(prefix + "_notifications")
	sql := "SELECT * FROM " + tableName + " WHERE closed = false AND id > ? AND recipient_id IN (?)"
	return GetAll(sql, -1, lastNotificationID, userIDs)
}
