package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) CfProjectDataInit() error {

	var fields []map[string]string
	if p.BlockData != nil && p.BlockData.BlockId < 134261 {
		fields = []map[string]string{{"project_id": "int64"}, {"lang_id": "int64"}, {"blurb_img": "string"}, {"head_img": "string"}, {"description_img": "string"}, {"picture": "string"}, {"video_type": "string"}, {"video_url_id": "string"}, {"news_img": "string"}, {"links": "string"}, {"sign": "bytes"}}
	} else {
		fields = []map[string]string{{"project_id": "int64"}, {"lang_id": "int64"}, {"blurb_img": "string"}, {"head_img": "string"}, {"description_img": "string"}, {"picture": "string"}, {"video_type": "string"}, {"video_url_id": "string"}, {"news_img": "string"}, {"links": "string"}, {"hide": "int64"}, {"sign": "bytes"}}
	}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CfProjectDataFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"project_id": "int", "lang_id": "tinyint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !utils.CheckInputData(p.TxMaps.String["blurb_img"], "img_url") && p.TxMaps.String["blurb_img"] != "0" {
		return fmt.Errorf("incorrect blurb_img")
	}
	if !utils.CheckInputData(p.TxMaps.String["head_img"], "img_url") && p.TxMaps.String["head_img"] != "0" {
		return fmt.Errorf("incorrect head_img")
	}
	if !utils.CheckInputData(p.TxMaps.String["description_img"], "img_url") && p.TxMaps.String["description_img"] != "0" {
		return fmt.Errorf("incorrect description_img")
	}
	if !utils.CheckInputData(p.TxMaps.String["picture"], "img_url") && p.TxMaps.String["picture"] != "0" {
		return fmt.Errorf("incorrect picture")
	}
	if !utils.CheckInputData(p.TxMaps.String["video_type"], "video_type") && p.TxMaps.String["video_type"] != "0" {
		return fmt.Errorf("incorrect blurb_img")
	}
	if !utils.CheckInputData(p.TxMaps.String["video_url_id"], "video_url_id") && p.TxMaps.String["video_url_id"] != "0" {
		return fmt.Errorf("incorrect video_url_id")
	}
	if !utils.CheckInputData(p.TxMaps.String["news_img"], "img_url") && p.TxMaps.String["news_img"] != "0" {
		return fmt.Errorf("incorrect news_img")
	}
	if !utils.CheckInputData(p.TxMaps.String["links"], "cf_links") && p.TxMaps.String["links"] != "0" {
		return fmt.Errorf("incorrect links")
	}

	// для подписи
	if p.BlockData != nil && p.BlockData.BlockId < 134261 {
		p.TxMap["hide"] = []byte(`0`)
	} else {
		if !utils.CheckInputData(p.TxMap["hide"], "boolean") {
			return fmt.Errorf("incorrect hide")
		}
	}

	// является ли юзер владельцем данного проекта и есть ли вообще такой проект
	projectUserId, err := p.Single("SELECT user_id FROM cf_projects WHERE user_id  =  ? AND id  =  ?", p.TxUserID, p.TxMaps.Int64["project_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if projectUserId == 0 {
		return fmt.Errorf("incorrect project_user_id")
	}

	var forSign string
	if p.BlockData != nil && p.BlockData.BlockId < 134261 {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["project_id"], p.TxMap["lang_id"], p.TxMap["blurb_img"], p.TxMap["head_img"], p.TxMap["description_img"], p.TxMap["picture"], p.TxMap["video_type"], p.TxMap["video_url_id"], p.TxMap["news_img"], p.TxMap["links"])
	} else {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["project_id"], p.TxMap["lang_id"], p.TxMap["blurb_img"], p.TxMap["head_img"], p.TxMap["description_img"], p.TxMap["picture"], p.TxMap["video_type"], p.TxMap["video_url_id"], p.TxMap["news_img"], p.TxMap["links"], p.TxMap["hide"])

	}

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

func (p *Parser) CfProjectData() error {

	// если описания проекта на этом языке еще нет, то добавляем, иначе - логируем и обновляем
	return p.selectiveLoggingAndUpd([]string{"blurb_img", "head_img", "description_img", "picture", "video_type", "video_url_id", "news_img", "links", "hide"}, []interface{}{p.TxMaps.String["blurb_img"], p.TxMaps.String["head_img"], p.TxMaps.String["description_img"], p.TxMaps.String["picture"], p.TxMaps.String["video_type"], p.TxMaps.String["video_url_id"], p.TxMaps.String["news_img"], p.TxMaps.String["links"], p.TxMaps.Int64["hide"]}, "cf_projects_data", []string{"project_id", "lang_id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["project_id"]), utils.Int64ToStr(p.TxMaps.Int64["lang_id"])})
}

func (p *Parser) CfProjectDataRollback() error {
	return p.selectiveRollback([]string{"blurb_img", "head_img", "description_img", "picture", "video_type", "video_url_id", "news_img", "links", "hide"}, "cf_projects_data", "project_id="+utils.Int64ToStr(p.TxMaps.Int64["project_id"])+" AND lang_id = "+utils.Int64ToStr(p.TxMaps.Int64["lang_id"]), true)
}

func (p *Parser) CfProjectDataRollbackFront() error {
	return p.limitRequestsRollback("cf_project_data")
}
