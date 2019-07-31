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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const rollbackHistoryLimit = 100

type historyResult struct {
	List []map[string]string `json:"list"`
}

func getHistoryHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)
	client := getClient(r)

	table := client.Prefix() + "_" + params["name"]
	rollbackTx := &model.RollbackTx{}
	txs, err := rollbackTx.GetRollbackTxsByTableIDAndTableName(params["id"], table, rollbackHistoryLimit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("rollback history")
		errorResponse(w, err)
		return
	}
	rollbackList := []map[string]string{}
	for _, tx := range *txs {
		if tx.Data == "" {
			continue
		}
		rollback := map[string]string{}
		if err := json.Unmarshal([]byte(tx.Data), &rollback); err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling rollbackTx.Data from JSON")
			errorResponse(w, err)
			return
		}
		rollbackList = append(rollbackList, rollback)
	}

	jsonResponse(w, &historyResult{rollbackList})
}
