// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

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

type keyInfoResult struct {
	Ecosystem string     `json:"ecosystem"`
	Name      string     `json:"name"`
	Roles     []roleInfo `json:"roles,omitempty"`
}

func getKeyInfoHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	keysList := make([]keyInfoResult, 0)
	keyID := converter.StringToAddress(params["wallet"])
	if keyID == 0 {
		errorResponse(w, errInvalidWallet.Errorf(params["wallet"]))
		return
	}

	ids, names, err := model.GetAllSystemStatesIDs()
	if err != nil {
		errorResponse(w, err)
		return
	}

	var found bool
	for i, ecosystemID := range ids {
		found, err = getEcosystemKey(keyID, ecosystemID)
		if err != nil {
			errorResponse(w, err)
			return
		}
		if !found {
			continue
		}
		keyRes := keyInfoResult{
			Ecosystem: converter.Int64ToStr(ecosystemID),
			Name:      names[i],
		}
		ra := &model.RolesParticipants{}
		roles, err := ra.SetTablePrefix(ecosystemID).GetActiveMemberRoles(keyID)
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
			} else {
				keyRes.Roles = append(keyRes.Roles, role)
			}
		}
		keysList = append(keysList, keyRes)
	}

	if len(keysList) == 0 && syspar.IsTestMode() {
		keysList = append(keysList, keyInfoResult{
			Ecosystem: converter.Int64ToStr(ids[0]),
			Name:      names[0],
		})
	}

	jsonResponse(w, &keysList)
}

func getEcosystemKey(keyID, ecosystemID int64) (bool, error) {
	// registration for the first ecosystem is open in test mode
	if ecosystemID == 1 && syspar.IsTestMode() {
		return true, nil
	}

	key := &model.Key{}
	key.SetTablePrefix(ecosystemID)
	return key.Get(keyID)
}
