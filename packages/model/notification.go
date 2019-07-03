// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

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
	RecipientID int64 `gorm:"recipient_id"`
	RoleID      int64 `gorm:"role_id"`
	Count       int64 `gorm:"count"`
}

// GetNotificationsCount returns all unclosed notifications by users and ecosystem through role_id
// if userIDs is nil or empty then filter will be skipped
func GetNotificationsCount(ecosystemID int64, accounts []string) ([]NotificationsCount, error) {
	result := make([]NotificationsCount, 0, len(accounts))
	for _, account := range accounts {
		roles, err := GetMemberRoles(nil, ecosystemID, account)
		if err != nil {
			return nil, err
		}
		roleList := make([]string, 0, len(roles))
		for _, role := range roles {
			roleList = append(roleList, converter.Int64ToStr(role))
		}

		query := `SELECT k.id as "recipient_id", recipient->>'role_id' as "role_id", count(*) as "count"
			FROM "1_notifications" n
			INNER JOIN "1_keys" k ON k.ecosystem = n.ecosystem AND k.account = ?
			WHERE n.ecosystem=? AND n.closed = 0 AND ((n.notification->>'type' = '1' and n.recipient->>'account' = ? ) or
				(n.notification->>'type' = '2' and (n.recipient->>'role_id' IN (?) and 
				(n.date_start_processing = 0 or n.processing_info->>'account' = ?))))
			GROUP BY recipient_id, role_id`

		list := make([]NotificationsCount, 0)
		err = GetDB(nil).Raw(query, account, ecosystemID, account, roleList, account).Scan(&list).Error
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
