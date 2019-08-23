// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package model

import (
	"fmt"

	"github.com/AplaProject/go-apla/packages/converter"
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

type NotificationsCount struct {
	RecipientID int64  `gorm:"recipient_id"`
	Account     string `gorm:"account"`
	RoleID      int64  `gorm:"role_id"`
	Count       int64  `gorm:"count"`
}

// GetNotificationsCount returns all unclosed notifications by users and ecosystem through role_id
// if userIDs is nil or empty then filter will be skipped
func GetNotificationsCount(ecosystemID int64, accounts []string) ([]NotificationsCount, error) {
	result := make([]NotificationsCount, 0, len(accounts))
	for _, account := range accounts {
		query := `SELECT k.id as "recipient_id", '0' as "role_id", count(n.id), k.account
			FROM "1_keys" k
			LEFT JOIN "1_notifications" n ON n.ecosystem = k.ecosystem AND n.closed = 0 AND n.notification->>'type' = '1' and n.recipient->>'account' = k.account
			WHERE k.ecosystem = ? AND k.account = ?
			GROUP BY recipient_id, k.account, role_id
			UNION
			SELECT k.id as "recipient_id", rp.role->>'id' as "role_id", count(n.id), k.account
			FROM "1_keys" k
			INNER JOIN "1_roles_participants" rp ON rp.member->>'account' = k.account
			LEFT JOIN "1_notifications" n ON n.ecosystem = k.ecosystem AND n.closed = 0 AND n.notification->>'type' = '2' AND n.recipient->>'role_id' = rp.role->>'id'
													AND (n.date_start_processing = 0 OR n.processing_info->>'account' = k.account)
			WHERE k.ecosystem=? AND k.account = ?
			GROUP BY recipient_id, k.account, role_id`

		list := make([]NotificationsCount, 0)
		err := GetDB(nil).Raw(query, ecosystemID, account, ecosystemID, account).Scan(&list).Error
		if err != nil {
			return nil, err
		}
		result = append(result, list...)
	}
	return result, nil
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
