package dcparser

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) AbusesInit() error {
	fields := []map[string]string{{"abuses": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AbusesFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	var abuses map[string]string
	err = json.Unmarshal(p.TxMap["abuses"], &abuses)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(abuses) > 100 {
		return fmt.Errorf(">100")
	}
	for userId, comment := range abuses {
		if !utils.CheckInputData(userId, "user_id") {
			return fmt.Errorf("incorrect abuses user_id")
		}
		if !utils.CheckInputData(comment, "abuse_comment") {
			return fmt.Errorf("incorrect abuse_comment")
		}
		// является ли данный юзер майнером
		err = p.checkMiner(utils.StrToInt64(userId))
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["abuses"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(p.Variables.Int64["limit_abuses"], "abuses", p.Variables.Int64["limit_abuses_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) Abuses() error {

	var abuses map[string]string
	err := json.Unmarshal(p.TxMap["abuses"], &abuses)
	if err != nil {
		return p.ErrInfo(err)
	}
	for userId, comment := range abuses {
		err = p.ExecSql("INSERT INTO abuses ( user_id, from_user_id, comment, time ) VALUES ( ?, ?, ?, ? )", userId, p.TxUserID, comment, p.BlockData.Time)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) AbusesRollback() error {
	var abuses map[string]string
	err := json.Unmarshal(p.TxMap["abuses"], &abuses)
	if err != nil {
		return p.ErrInfo(err)
	}
	for userId, _ := range abuses {
		err = p.ExecSql("DELETE FROM abuses WHERE user_id = ? AND from_user_id = ? AND time = ?", userId, p.TxUserID, p.BlockData.Time)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) AbusesRollbackFront() error {
	return p.limitRequestsRollback("abuses")
}
