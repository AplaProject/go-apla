package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"log"
	"encoding/json"
	//"regexp"
	//"math"
	//"strings"
	//	"os"
	//"time"
	//"strings"
	//"bytes"
	//"github.com/DayLightProject/go-daylight/packages/consts"
	//"math"
	//"database/sql"
	"strings"
)

func (p *Parser) ChangeCommissionInit() error {

	fields := []map[string]string{{"commission": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeCommissionFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	if len(p.TxMap["commission"]) > 3000 {
		return p.ErrInfo("len commission")
	}
	commissionMap := make(map[string][3]float64)
	err = json.Unmarshal(p.TxMaps.Bytes["commission"], &commissionMap)
	if err != nil {
		return p.ErrInfo(err)
	}
	var currencyArray []int64
	minusCf := 0
	for currencyId_, data := range commissionMap {

		currencyId := utils.StrToInt64(currencyId_)
		if len(data) != 3 {
			return p.ErrInfo("len(data) !=3")
		}
		if !utils.CheckInputData(currencyId, "int") {
			return p.ErrInfo("currencyId")
		}
		// % от 0 до 10
		if !utils.CheckInputData(utils.Float64ToStrPct(data[0]), "currency_commission") || data[0] > 10 {
			return p.ErrInfo("pct")
		}
		// минимальная комиссия от 0. При 0% будет = 0
		if !utils.CheckInputData(utils.Float64ToStrPct(data[1]), "currency_commission") {
			return p.ErrInfo("currency_min_commission")
		}
		// макс. комиссия. 0 - значит, считается по %
		if !utils.CheckInputData(utils.Float64ToStrPct(data[2]), "currency_commission") {
			return p.ErrInfo("currency_max_commission")
		}
		if data[1] > data[2] && data[2] > 0 {
			return p.ErrInfo("currency_max_commission")
		}
		// проверим, существует ли такая валюта в таблице DC-валют
		if ok, err := p.CheckCurrency(currencyId); !ok {
			// если нет, то это может быть $currency_id 1000, которая определяет комиссию для всх CF-валют
			if currencyId != 1000 {
				return p.ErrInfo(err)
			}
		}
		if currencyId != 1000 {
			currencyArray = append(currencyArray, currencyId)
		} else {
			minusCf = 1
		}
	}

	count, err := p.Single("SELECT count(id) FROM currency WHERE id IN (" + strings.Join(utils.SliceInt64ToString(currencyArray), ",") + ")").Int()
	if err != nil {
		return p.ErrInfo(err)
	}
	if count != len(commissionMap)-minusCf {
		return p.ErrInfo("count != len(commissionMap) - minusCf")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["commission"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil || !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(p.Variables.Int64["limit_commission"], "commission", p.Variables.Int64["limit_commission_period"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) ChangeCommission() error {

	logData, err := p.OneRow("SELECT * FROM commission WHERE user_id  =  ?", p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	// если есть, что логировать, то логируем
	if len(logData) > 0 {
		logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_commission ( commission, block_id, prev_log_id ) VALUES ( ?, ?, ? )", "log_id", logData["commission"], p.BlockData.BlockId, logData["log_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE commission SET commission = ?, log_id = ? WHERE user_id = ?", p.TxMaps.Bytes["commission"], logId, p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		err = p.ExecSql("INSERT INTO commission ( user_id, commission ) VALUES ( ?, ? )", p.TxUserID, p.TxMaps.Bytes["commission"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) ChangeCommissionRollback() error {
	return p.generalRollback("commission", p.TxUserID, "", false)
}

func (p *Parser) ChangeCommissionRollbackFront() error {
	return p.limitRequestsRollback("commission")
}
