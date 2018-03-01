// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
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
