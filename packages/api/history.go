package api

import (
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	log "github.com/sirupsen/logrus"
)

const rollbackHistoryLimit = 100

type historyResult struct {
	List []map[string]string `json:"list"`
}

func getHistory(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	table := getPrefix(data) + "_" + data.params["table"].(string)
	id := converter.StrToInt64(data.params["id"].(string))

	rbID, err := model.GetRollbackIDForTableRow(table, id)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("rollback id for table row")
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	result := historyResult{}
	result.List, err = model.GetRollbackHistory(rbID, rollbackHistoryLimit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("rollback history")
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	data.result = &result
	return nil
}
