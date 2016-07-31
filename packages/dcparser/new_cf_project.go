package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

func (p *Parser) NewCfProjectInit() error {

	fields := []map[string]string{{"currency_id": "int64"}, {"amount": "int64"}, {"end_time": "int64"}, {"latitude": "float64"}, {"longitude": "float64"}, {"category_id": "int64"}, {"project_currency_name": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewCfProjectFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"currency_id": "int", "amount": "int", "end_time": "int", "latitude": "coordinate", "longitude": "coordinate", "category_id": "smallint", "project_currency_name": "cf_currency_name"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.BlockData == nil || p.BlockData.BlockId >= 168904 {
		// является ли данный юзер майнером
		err = p.checkMiner(p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	var time1, time2 int64
	if p.BlockData != nil {
		time1 = p.BlockData.Time
		time2 = time1
	} else { // голая тр-ия с запасом 30 сек на время генерации блока. Т.к. при попадинии в блок время будет уже другим
		time1 = time.Now().Unix() - 30
		time2 = time.Now().Unix() + 30
	}

	// дата завершения проекта не может быть более чем на 91 дней и менее чем на 6 дней вперед от текущего времени
	if p.TxMaps.Int64["end_time"]-time1 > 3600*24*91 || p.TxMaps.Int64["end_time"]-time2 < 3600*24*6 {
		return p.ErrInfo("incorrect end_time")
	}
	if ok, err := p.CheckCurrency(p.TxMaps.Int64["currency_id"]); !ok {
		return p.ErrInfo(err)
	}

	if p.TxMaps.Int64["amount"] <= 0 {
		return p.ErrInfo("amount<=0")
	}

	// проверим, не занято ли имя валюты
	currency, err := p.Single("SELECT id FROM cf_projects WHERE project_currency_name  =  ? AND close_block_id  =  0 AND del_block_id  =  0", p.TxMaps.String["project_currency_name"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if currency != 0 {
		return p.ErrInfo("exists project_currency_name'")
	}

	// проверим, не занято ли имя валюты
	currency, err = p.Single("SELECT id FROM cf_currency WHERE name  =  ?", p.TxMaps.String["project_currency_name"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if currency != 0 {
		return p.ErrInfo("exists cf_currency")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["currency_id"], p.TxMap["amount"], p.TxMap["end_time"], p.TxMap["latitude"], p.TxMap["longitude"], p.TxMap["category_id"], p.TxMap["project_currency_name"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_NEW_CF_PROJECT, "new_cf_project", consts.LIMIT_NEW_CF_PROJECT_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewCfProject() error {

	err := p.ExecSql("INSERT INTO cf_projects ( user_id, currency_id, amount, project_currency_name, start_time, end_time, latitude, longitude, category_id, block_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )", p.TxUserID, p.TxMaps.Int64["currency_id"], p.TxMaps.Int64["amount"], p.TxMaps.String["project_currency_name"], p.BlockData.Time, p.TxMaps.Int64["end_time"], p.TxMaps.Float64["latitude"], p.TxMaps.Float64["longitude"], p.TxMaps.Int64["category_id"], p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewCfProjectRollback() error {
	err := p.ExecSql("DELETE FROM cf_projects WHERE block_id = ? AND user_id = ?", p.BlockData.BlockId, p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("cf_projects", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewCfProjectRollbackFront() error {
	return p.limitRequestsRollback("new_cf_project")
}
