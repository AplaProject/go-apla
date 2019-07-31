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

package notificator

import (
	"encoding/json"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/publisher"

	log "github.com/sirupsen/logrus"
)

type notificationRecord struct {
	EcosystemID  string `json:"ecosystem"`
	RoleID       string `json:"role_id"`
	RecordsCount int64  `json:"count"`
}

// UpdateNotifications send stats about unreaded messages to centrifugo for ecosystem
func UpdateNotifications(ecosystemID int64, accounts []string) {
	notificationsStats, err := getEcosystemNotificationStats(ecosystemID, accounts)
	if err != nil {
		return
	}

	for user, n := range notificationsStats {
		sendUserStats(user, *n)
	}
}

// UpdateRolesNotifications send stats about unreaded messages to centrifugo for ecosystem
func UpdateRolesNotifications(ecosystemID int64, roles []int64) {
	members, _ := model.GetRoleMembers(nil, ecosystemID, roles)
	UpdateNotifications(ecosystemID, members)
}

func getEcosystemNotificationStats(ecosystemID int64, users []string) (map[int64]*[]notificationRecord, error) {
	result, err := model.GetNotificationsCount(ecosystemID, users)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting notification count")
		return nil, err
	}

	return parseRecipientNotification(result, ecosystemID), nil
}

func parseRecipientNotification(rows []model.NotificationsCount, systemID int64) map[int64]*[]notificationRecord {
	recipientNotifications := make(map[int64]*[]notificationRecord)

	for _, r := range rows {
		if r.RecipientID == 0 {
			continue
		}

		roleNotifications := notificationRecord{
			EcosystemID:  converter.Int64ToStr(systemID),
			RoleID:       converter.Int64ToStr(r.RoleID),
			RecordsCount: r.Count,
		}

		nr, ok := recipientNotifications[r.RecipientID]
		if ok {
			*nr = append(*nr, roleNotifications)
			continue
		}

		records := []notificationRecord{
			roleNotifications,
		}

		recipientNotifications[r.RecipientID] = &records
	}

	return recipientNotifications
}

func sendUserStats(user int64, stats []notificationRecord) {
	rawStats, err := json.Marshal(stats)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("notification statistic")
	}

	ok, err := publisher.Write(user, string(rawStats))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing to centrifugo")
	}

	if !ok {
		log.WithFields(log.Fields{"type": consts.CentrifugoError, "error": err}).Error("writing to centrifugo")
	}
}
