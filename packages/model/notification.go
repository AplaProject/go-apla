package model

import (
	"strconv"
)

const notificationTableSuffix = "_notifications"

// Notification structure
type Notification struct {
	tableName              string
	ID                     int64 `gorm:"primary_key;not null"`
	StartedProcessingTime  int64
	StartedTime            int64
	BodyText               string
	RecipientID            int64
	RecipientName          string `gorm:"size:255"`
	StartedProcessingID    int64
	Name                   string `gorm:"size:255"`
	RoleID                 int64
	RoleName               string `gorm:"size:255"`
	PageValInt             int64
	PageValStr             string `gorm:"size:255"`
	Closed                 bool
	RecipientAvatar        string
	NotificationType       int64
	FinishedProcessingID   int64
	FinishedProcessingTime int64
	PageName               string `gorm:"size:255"`
}

// SetTablePrefix set table Prefix
func (n *Notification) SetTablePrefix(tablePrefix string) {
	n.tableName = tablePrefix + notificationTableSuffix
}

// TableName returns table name
func (n *Notification) TableName() string {
	return n.tableName
}

// GetNotificationsCount returns all unclosed notifications by users and ecosystem through role_id
// if userIDs is nil or empty then filter will be skipped
func GetNotificationsCount(ecosystemID int64, userIDs []int64) ([]map[string]string, error) {
	filter, params := getNotificationCountFilter(userIDs)
	query := `SELECT recipient_id, role_id, count(*) cnt 
	FROM "` + strconv.FormatInt(ecosystemID, 10) + notificationTableSuffix + `" 
	` + filter + ` 
	GROUP BY "recipient_id", "role_id"`

	return GetAllTransaction(nil, query, -1, params...)
}

func getNotificationCountFilter(users []int64) (filter string, params []interface{}) {
	filter = ` WHERE closed = 0 `

	if len(users) > 0 {
		filter += `AND recipient_id IN (?) `
		params = append(params, users)
	}

	return
}
