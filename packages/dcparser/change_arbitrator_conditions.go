package dcparser

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

func (p *Parser) ChangeArbitratorConditionsInit() error {

	fields := []map[string]string{{"conditions": "string"}, {"url": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeArbitratorConditionsFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	if len(p.TxMap["conditions"]) > 3000 {
		return fmt.Errorf("incorrect conditions")
	}

	// может прийти [0]
	if p.TxMaps.String["conditions"] != "[0]" {

		var conditions map[string][5]string
		err = json.Unmarshal(p.TxMap["conditions"], &conditions)
		if err != nil {
			return p.ErrInfo(err)
		}
		if len(conditions) == 0 {
			return fmt.Errorf("len(conditions) == 0")
		}

		var currencyArray []string
		var minusCf int64
		for currencyId, data := range conditions {
			if !utils.CheckInputData(currencyId, "bigint") {
				return fmt.Errorf("incorrect currencyId")
			}
			if len(data) != 5 {
				return fmt.Errorf("incorrect data")
			}
			minAmount := utils.StrToFloat64(data[0])
			maxAmount := utils.StrToFloat64(data[1])
			minCommission := utils.StrToFloat64(data[2])
			maxCommission := utils.StrToFloat64(data[3])
			commissionPct := utils.StrToFloat64(data[4])

			if !utils.CheckInputData(data[0], "amount") || minAmount < 0.01 {
				return fmt.Errorf("incorrect minAmount")
			}
			if !utils.CheckInputData(data[1], "amount") {
				return fmt.Errorf("incorrect maxAmount")
			}
			if !utils.CheckInputData(data[2], "amount") || minCommission < 0.01 {
				return fmt.Errorf("incorrect minCommission")
			}
			if !utils.CheckInputData(data[3], "amount") {
				return fmt.Errorf("incorrect maxCommission")
			}
			if !utils.CheckInputData(data[4], "pct") || commissionPct > 10 || commissionPct < 0.01 {
				return fmt.Errorf("incorrect commissionPct")
			}
			if maxCommission > 0 && minCommission > maxCommission {
				return fmt.Errorf("minCommission > maxCommission")
			}
			if maxAmount > 0 && minAmount > maxAmount {
				return fmt.Errorf("minAmount > maxAmount")
			}
			// проверим, существует ли такая валюта в таблице DC-валют
			if ok, err := p.CheckCurrency(utils.StrToInt64(currencyId)); !ok {
				// если нет, то это может быть $currency_id 1000, которая определяет комиссию для всх CF-валют
				if currencyId != "1000" {
					return p.ErrInfo(err)
				}
			}
			if currencyId != "1000" {
				currencyArray = append(currencyArray, currencyId)
			} else {
				minusCf = 1
			}
		}
		count, err := p.Single("SELECT count(id) FROM currency WHERE id IN (" + strings.Join(currencyArray, ",") + ")").Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if count != int64(len(conditions))-minusCf {
			return p.ErrInfo("count != int64(len(conditions)) - minusCf")
		}
	}

	if !utils.CheckInputData(p.TxMaps.String["url"], "arbitrator_url") && p.TxMaps.String["url"] != "0" {
		return fmt.Errorf("incorrect url")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["conditions"], p.TxMap["url"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_CHANGE_ARBITRATOR_CONDITIONS, "change_arbitrator_conditions", consts.LIMIT_CHANGE_ARBITRATOR_CONDITIONS_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeArbitratorConditions() error {

	logData, err := p.OneRow("SELECT * FROM arbitrator_conditions WHERE user_id  =  ?", p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	// если есть, что логировать, то логируем
	if len(logData) > 0 {
		logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_arbitrator_conditions ( conditions, block_id, prev_log_id ) VALUES ( ?, ?, ? )", "log_id", logData["conditions"], p.BlockData.BlockId, logData["log_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE arbitrator_conditions SET conditions = ?, log_id = ? WHERE user_id = ?", p.TxMaps.String["conditions"], logId, p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		err = p.ExecSql("INSERT INTO arbitrator_conditions ( user_id, conditions ) VALUES ( ?, ? )", p.TxUserID, p.TxMaps.String["conditions"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	err = p.selectiveLoggingAndUpd([]string{"url"}, []interface{}{p.TxMaps.String["url"]}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) ChangeArbitratorConditionsRollback() error {
	err := p.selectiveRollback([]string{"url"}, "users", "user_id="+utils.Int64ToStr(p.TxUserID), false)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.generalRollback("arbitrator_conditions", p.TxUserID, "", false)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeArbitratorConditionsRollbackFront() error {
	return p.limitRequestsRollback("change_arbitrator_conditions")
}
