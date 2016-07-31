package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangeKeyCloseInit() error {

	fields := []map[string]string{{"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeKeyCloseFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, не стоит ли уже close
	changeKeyClose, err := p.Single("SELECT change_key_close FROM users WHERE user_id  =  ?", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if changeKeyClose > 0 {
		return p.ErrInfo("change_key_close=1")
	}

	forSign := fmt.Sprintf("%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) ChangeKeyClose() error {
	return p.selectiveLoggingAndUpd([]string{"change_key_close"}, []interface{}{"1"}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
}

func (p *Parser) ChangeKeyCloseRollback() error {
	return p.selectiveRollback([]string{"change_key_close"}, "users", "user_id="+utils.Int64ToStr(p.TxUserID), false)
}

func (p *Parser) ChangeKeyCloseRollbackFront() error {
	return nil
}
