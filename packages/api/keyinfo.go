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

package api

import (
	"encoding/json"
	"net/http"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type roleInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type notifyInfo struct {
	RoleID string `json:"role_id"`
	Count  int64  `json:"count"`
}

type keyInfoResult struct {
	Account    string              `json:"account"`
	Ecosystems []*keyEcosystemInfo `json:"ecosystems"`
}

type keyEcosystemInfo struct {
	Ecosystem     string       `json:"ecosystem"`
	Name          string       `json:"name"`
	Roles         []roleInfo   `json:"roles,omitempty"`
	Notifications []notifyInfo `json:"notifications,omitempty"`
}

func (m Mode) getKeyInfoHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	keysList := make([]*keyEcosystemInfo, 0)
	keyID := converter.StringToAddress(params["wallet"])
	if keyID == 0 {
		errorResponse(w, errInvalidWallet.Errorf(params["wallet"]))
		return
	}

	ids, names, err := m.EcosysLookupGetter.GetEcosystemLookup()
	if err != nil {
		errorResponse(w, err)
		return
	}

	var (
		account string
		found   bool
	)

	for i, ecosystemID := range ids {
		key := &model.Key{}
		key.SetTablePrefix(ecosystemID)
		found, err = key.Get(keyID)
		if err != nil {
			errorResponse(w, err)
			return
		}
		if !found {
			continue
		}

		// TODO: delete after switching to another account storage scheme
		if len(account) == 0 {
			account = key.AccountID
		}

		keyRes := &keyEcosystemInfo{
			Ecosystem: converter.Int64ToStr(ecosystemID),
			Name:      names[i],
		}
		ra := &model.RolesParticipants{}
		roles, err := ra.SetTablePrefix(ecosystemID).GetActiveMemberRoles(key.AccountID)
		if err != nil {
			errorResponse(w, err)
			return
		}
		for _, r := range roles {
			var role roleInfo
			if err := json.Unmarshal([]byte(r.Role), &role); err != nil {
				logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling role")
				errorResponse(w, err)
				return
			}
			keyRes.Roles = append(keyRes.Roles, role)
		}
		keyRes.Notifications, err = m.getNotifications(ecosystemID, key)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting notifications")
			errorResponse(w, err)
			return
		}

		keysList = append(keysList, keyRes)
	}

	// in test mode, registration is open in the first ecosystem
	if len(keysList) == 0 && syspar.IsTestMode() {
		account = converter.AddressToString(keyID)
		keysList = append(keysList, &keyEcosystemInfo{
			Ecosystem: converter.Int64ToStr(ids[0]),
			Name:      names[0],
		})
	}

	jsonResponse(w, &keyInfoResult{
		Account:    account,
		Ecosystems: keysList,
	})
}

func (m Mode) getNotifications(ecosystemID int64, key *model.Key) ([]notifyInfo, error) {
	notif, err := model.GetNotificationsCount(ecosystemID, []string{key.AccountID})
	if err != nil {
		return nil, err
	}

	list := make([]notifyInfo, 0)
	for _, n := range notif {
		if n.RecipientID != key.ID {
			continue
		}

		list = append(list, notifyInfo{
			RoleID: converter.Int64ToStr(n.RoleID),
			Count:  n.Count,
		})
	}
	return list, nil
}
