// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package model

import (
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/converter"
)

const (
	notificationTableSuffix = "_notifications"

	NotificationTypeSingle = 1
	NotificationTypeRole   = 2
)

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
	query := `SELECT recipient->>'member_id' "recipient_id", recipient->>'role_id' "role_id", count(*) cnt
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
