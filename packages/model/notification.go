package model

import (
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/converter"
)

const (
	notificationTableSuffix = "_notifications"

	NotificationTypeSingle = 1
	NotificationTypeRole   = 2
)

// Notification structure
type Notification struct {
	ecosystem           int64
	ID                  int64  `gorm:"primary_key;not null"`
	Recipient           string `gorm:"type:jsonb(PostgreSQL)`
	Sender              string `gorm:"type:jsonb(PostgreSQL)`
	Notification        string `gorm:"type:jsonb(PostgreSQL)`
	PageParams          string `gorm:"type:jsonb(PostgreSQL)`
	ProcessingInfo      string `gorm:"type:jsonb(PostgreSQL)`
	PageName            string `gorm:"size:255"`
	DateCreated         int64
	DateStartProcessing int64
	DateClosed          int64
	Closed              bool
}

// SetTablePrefix set table Prefix
func (n *Notification) SetTablePrefix(tablePrefix string) {
	n.ecosystem = converter.StrToInt64(tablePrefix)
}

// TableName returns table name
func (n *Notification) TableName() string {
	if n.ecosystem == 0 {
		n.ecosystem = 1
	}
	return `1_notifications`
}

// GetNotificationsCount returns all unclosed notifications by users and ecosystem through role_id
// if userIDs is nil or empty then filter will be skipped
func GetNotificationsCount(ecosystemID int64, userIDs []int64) ([]map[string]string, error) {
	filter, params := getNotificationCountFilter(userIDs, ecosystemID)
	query := `SELECT recipient->>'member_id' "recipient_id", recipient->>'role_id' "role_id", count(*) cnt
	FROM "1_notifications" 
	` + filter + ` 
	GROUP BY 1,2`

	return GetAllTransaction(nil, query, -1, params...)
}

func getNotificationCountFilter(users []int64, ecosystemID int64) (filter string, params []interface{}) {
	filter = fmt.Sprintf(` WHERE closed = 0 and ecosystem = '%d' `, ecosystemID)

	if len(users) > 0 {
		filter += `AND recipient->>'member_id' IN (?) `
		usersStrs := []string{}
		for _, user := range users {
			usersStrs = append(usersStrs, converter.Int64ToStr(user))
		}
		params = append(params, usersStrs)
	}

	return
}
