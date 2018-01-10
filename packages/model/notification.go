package model

import (
	"bytes"
	"strconv"
)

const notificationTableSuffix = "_notifications"

// Notification structure
type Notification struct {
	tableName              string
	ID                     uint64 `gorm:"primary_key;not null;size:255"`
	StartedProcessingTime  int64
	StartedTime            int64
	BodyText               string
	RecipientID            uint64
	RecipientName          string `gorm:"size:255"`
	StartedProcessingID    uint64
	Name                   string `gorm:"size:255"`
	RoleID                 uint64
	RoleName               string `gorm:"size:255"`
	PageValInt             uint64
	PageValStr             string `gorm:"size:255"`
	Closed                 bool
	RecipientAvatar        string
	NotificationType       uint64
	FinishedProcessingID   uint64
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
	query := `SELECT recipient_id, role_id, count(*) cnt 
	 FROM "` + strconv.FormatInt(ecosystemID, 10) + `_notifications" 
	 WHERE closed = false ` + getNotificationCountFilter(userIDs) + ` 
	 GROUP BY "recipient_id", "role_id";`

	return GetAll(query, -1)
}

func getNotificationCountFilter(users []int64) string {
	if users == nil || len(users) == 0 {
		return ""
	}

	return ` AND recipient_id IN (` + IDList(users).SQLString() + `) `
}

// IDList represent extended []int64
type IDList []int64

// SQLString returns list of items separated by "," as string
func (l IDList) SQLString() string {
	var bts []byte
	buf := bytes.NewBuffer(bts)

	for i, item := range l {
		if i > 0 {
			buf.WriteString(",")
		}

		buf.WriteString(strconv.FormatInt(item, 10))
	}

	return buf.String()
}
