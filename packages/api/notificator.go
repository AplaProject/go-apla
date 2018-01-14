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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/notificator"

	log "github.com/sirupsen/logrus"
)

type idItem struct {
	ID          int64 `json:"id"`
	EcosystemID int64 `json:"ecosystem"`
}

type updateNotificatorResult struct {
	Result bool `json:"result"`
}

func updateNotificator(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var list []idItem

	err := json.Unmarshal([]byte(data.params["ids"].(string)), &list)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling ids")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	stateList := make(map[int64][]int64)
	for _, item := range list {
		if _, ok := stateList[item.EcosystemID]; !ok {
			stateList[item.EcosystemID] = make([]int64, 0)
		}
		stateList[item.EcosystemID] = append(stateList[item.EcosystemID], item.ID)
	}

	for ecosystemID, users := range stateList {
		notificator.UpdateNotifications(ecosystemID, users)
	}
	data.result = &updateNotificatorResult{Result: true}
	return nil
}
