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
	"github.com/AplaProject/go-apla/packages/model"
	log "github.com/sirupsen/logrus"
)

const rollbackHistoryLimit = 100

type historyResult struct {
	List []map[string]string `json:"list"`
}

func getHistory(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	table := getPrefix(data) + "_" + data.params["table"].(string)
	id := data.params["id"].(string)
	rollbackTx := &model.RollbackTx{}
	txs, err := rollbackTx.GetRollbackTxsByTableIDAndTableName(id, table, rollbackHistoryLimit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("rollback history")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	rollbackList := []map[string]string{}
	for _, tx := range *txs {
		if tx.Data == "" {
			continue
		}
		rollback := map[string]string{}
		if err := json.Unmarshal([]byte(tx.Data), &rollback); err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling rollbackTx.Data from JSON")
			return errorAPI(w, err, http.StatusInternalServerError)
		}
		rollbackList = append(rollbackList, rollback)
	}
	data.result = &historyResult{rollbackList}
	return nil
}
