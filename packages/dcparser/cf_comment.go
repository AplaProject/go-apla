package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

func (p *Parser) CfCommentInit() error {

	fields := []map[string]string{{"project_id": "int64"}, {"lang_id": "int64"}, {"comment": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CfCommentFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"project_id": "int", "comment": "cf_comment"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if !utils.CheckInputData(p.TxMaps.Int64["lang_id"], "tinyint") || p.TxMaps.Int64["lang_id"] <= 0 {
		return fmt.Errorf("incorrect lang_id")
	}

	var txTime int64
	if p.BlockData != nil {
		txTime = p.BlockData.Time
	} else { // голая тр-ия с запасом 30 сек на время генерации блока. Т.к. при попадинии в блок время будет уже другим
		txTime = time.Now().Unix() - 30
	}

	// автор проекта может писать по 1 комменту за каждую языковую версию
	author, err := p.Single("SELECT id FROM cf_projects WHERE user_id  =  ? AND id  =  ?", p.TxUserID, p.TxMaps.Int64["project_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	addSql := ""
	if author > 0 {
		addSql = " AND lang_id = " + utils.Int64ToStr(p.TxMaps.Int64["lang_id"])
	} else {
		addSql = ""
	}

	// проверим, есть ли у данного юзера другие комменты за данный проект
	commentTime, err := p.Single("SELECT max(time) FROM cf_comments WHERE user_id  =  ? AND project_id  =  ? "+addSql, p.TxUserID, p.TxMaps.Int64["project_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// в 1 проект можно писать только 1 комммент в сутки
	if txTime-commentTime < consts.LIMIT_TIME_COMMENTS_CF_PROJECT {
		return p.ErrInfo("comment_time")
	}

	// финансировал ли данный юзер этот проект
	funder, err := p.Single(" SELECT id FROM cf_funding WHERE user_id  =  ? AND project_id  =  ?", p.TxUserID, p.TxMaps.Int64["project_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if funder == 0 {
		//  или может быть он его автор
		author, err := p.Single("SELECT id FROM cf_projects WHERE user_id  =  ? AND id  =  ?", p.TxUserID, p.TxMaps.Int64["project_id"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if author == 0 {
			return p.ErrInfo("!funder || !author")
		}
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["project_id"], p.TxMap["lang_id"], p.TxMap["comment"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_CF_COMMENTS, "cf_comments", consts.LIMIT_CF_COMMENTS_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CfComment() error {

	err := p.ExecSql("INSERT INTO cf_comments ( user_id, project_id, lang_id, comment, time, block_id ) VALUES ( ?, ?, ?, ?, ?, ? )", p.TxUserID, p.TxMaps.Int64["project_id"], p.TxMaps.Int64["lang_id"], p.TxMaps.String["comment"], p.BlockData.Time, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) CfCommentRollback() error {
	err := p.ExecSql("DELETE FROM cf_comments WHERE block_id = ? AND user_id = ? AND project_id = ? AND lang_id = ?", p.BlockData.BlockId, p.TxUserID, p.TxMaps.Int64["project_id"], p.TxMaps.Int64["lang_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("cf_comments", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CfCommentRollbackFront() error {
	return p.limitRequestsRollback("cf_comments")
}
