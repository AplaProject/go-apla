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

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type roleInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type keyInfoResult struct {
	Ecosystem string     `json:"ecosystem"`
	Roles     []roleInfo `json:"roles,omitempty"`
}

func keyInfo(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {

	keysList := make([]keyInfoResult, 0)
	keyID := converter.StringToAddress(data.params[`wallet`].(string))
	if keyID == 0 {
		return errorAPI(w, `E_INVALIDWALLET`, http.StatusBadRequest, data.params[`wallet`].(string))
	}
	rows, err := model.DBConn.Table(`1_ecosystems`).Select("id").Rows()
	if err != nil {
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	var (
		id    string
		found bool
	)
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return errorAPI(w, err, http.StatusInternalServerError)
		}
		key := &model.Key{}
		ecosystemID := converter.StrToInt64(id)
		key.SetTablePrefix(ecosystemID)
		found, err = key.Get(keyID)
		if err != nil {
			return errorAPI(w, err, http.StatusInternalServerError)
		}
		if !found {
			continue
		}
		keyRes := keyInfoResult{Ecosystem: id}
		ra := &model.RolesParticipants{}
		roles, err := ra.SetTablePrefix(ecosystemID).GetActiveMemberRoles(keyID)
		if err != nil {
			return errorAPI(w, err, http.StatusInternalServerError)
		}
		for _, r := range roles {
			var role roleInfo
			if err := json.Unmarshal([]byte(r.Role), &role); err != nil {
				log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling role")
				return errorAPI(w, `E_SERVER`, http.StatusInternalServerError)
			} else {
				keyRes.Roles = append(keyRes.Roles, role)
			}
		}
		keysList = append(keysList, keyRes)
	}
	data.result = &keysList
	return
}
