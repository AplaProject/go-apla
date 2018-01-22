package querycost

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	log "github.com/sirupsen/logrus"
)

// explainQueryCost is counting query execution time
func explainQueryCost(transaction *model.DbTransaction, withAnalyze bool, query string, args ...interface{}) (int64, error) {
	var planStr string
	explainTpl := "EXPLAIN (FORMAT JSON) %s"
	if withAnalyze {
		explainTpl = "EXPLAIN ANALYZE (FORMAT JSON) %s"
	}
	err := model.GetDB(transaction).Raw(fmt.Sprintf(explainTpl, query), args...).Row().Scan(&planStr)
	switch {
	case err == sql.ErrNoRows:
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": query}).Error("no rows while explaining query")
		return 0, errors.New("No rows")
	case err != nil:
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": query}).Error("error explaining query")
		return 0, err
	}
	var queryPlan []map[string]interface{}
	dec := json.NewDecoder(strings.NewReader(planStr))
	dec.UseNumber()
	if err := dec.Decode(&queryPlan); err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("decoding query plan from JSON")
		return 0, err
	}
	if len(queryPlan) == 0 {
		log.Error("Query plan is empty")
		return 0, errors.New("Query plan is empty")
	}
	firstNode := queryPlan[0]
	var plan interface{}
	var ok bool
	if plan, ok = firstNode["Plan"]; !ok {
		log.Error("No Plan key in result")
		return 0, errors.New("No Plan key in result")
	}

	planMap, ok := plan.(map[string]interface{})
	if !ok {
		log.Error("Plan is not map[string]interface{}")
		return 0, errors.New("Plan is not map[string]interface{}")
	}

	totalCost, ok := planMap["Total Cost"]
	if !ok {
		return 0, errors.New("PlanMap has no TotalCost")
	}

	totalCostNum, ok := totalCost.(json.Number)
	if !ok {
		log.Error("PlanMap has no TotalCost")
		return 0, errors.New("Total cost is not a number")
	}

	totalCostF64, err := totalCostNum.Float64()
	if err != nil {
		log.Error("Total cost is not a number")
		return 0, err
	}
	return int64(totalCostF64), nil
}
