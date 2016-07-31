package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) CfProjectChangeCategoryInit() error {

	fields := []map[string]string{{"project_id": "int64"}, {"category_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CfProjectChangeCategoryFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"project_id": "int", "category_id": "tinyint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли юзер владельцем данного проекта и есть ли вообще такой проект
	projectUserId, err := p.Single("SELECT user_id FROM cf_projects WHERE user_id  =  ? AND id  =  ?", p.TxUserID, p.TxMaps.Int64["project_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if projectUserId == 0 {
		return p.ErrInfo("incorrect project_user_id")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["project_id"], p.TxMap["category_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_CF_PROJECT_DATA, "cf_project_data", consts.LIMIT_CF_PROJECT_DATA_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CfProjectChangeCategory() error {
	return p.selectiveLoggingAndUpd([]string{"category_id"}, []interface{}{p.TxMaps.Int64["category_id"]}, "cf_projects", []string{"id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["project_id"])})
}

func (p *Parser) CfProjectChangeCategoryRollback() error {
	return p.selectiveRollback([]string{"category_id"}, "cf_projects", "id="+utils.Int64ToStr(p.TxMaps.Int64["project_id"]), false)

}

func (p *Parser) CfProjectChangeCategoryRollbackFront() error {
	return p.limitRequestsRollback("cf_project_data")
}
