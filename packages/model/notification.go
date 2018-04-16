package model

import (
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/converter"
)

const notificationTableSuffix = "_notifications"

// Notification structure
type Notification struct {
	tableName           string
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
	query := `SELECT recipient->>'member_id', recipient->>'role_id', count(*) cnt 
	FROM "` + strconv.FormatInt(ecosystemID, 10) + notificationTableSuffix + `" 
	` + filter + ` 
	GROUP BY 1,2`

	return GetAllTransaction(nil, query, -1, params...)
}

func getNotificationCountFilter(users []int64) (filter string, params []interface{}) {
	filter = ` WHERE closed = 0 `

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
