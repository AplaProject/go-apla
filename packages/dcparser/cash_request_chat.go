package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"database/sql"
)

func (p *Parser) CashRequestChatInit() error {
	fields := []map[string]string{{"cash_request_id": "int64"}, {"message": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CashRequestChatFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"cash_request_id": "bigint", "message": "string"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}
	var to_user_id, user_id int64
	var status string
	var hash_code []byte
	err = p.QueryRow(p.FormatQuery("SELECT to_user_id, user_id, status, hash_code FROM cash_requests WHERE id  =  ?"), p.TxMaps.Int64["cash_request_id"]).Scan(&to_user_id, &user_id, &status, &hash_code)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}

	// ID cash_requests юзер указал сам, значит это может быть случайное число.
	// проверим, является получателем или отправителем наш юзер
	if to_user_id != p.TxUserID && user_id != p.TxUserID {
		return p.ErrInfo("to_user_id != p.TxUserID && user_id != p.TxUserID ")
	}
	// должно быть pending
	if status != "pending" {
		return p.ErrInfo("status!=pending")
	}
	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["cash_request_id"], p.TxMap["message"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	return nil
}
func (p *Parser) CashRequestChat() error {
	/*err := p.ExecSql("UPDATE auto_payments SET del_block_id = ? WHERE id = ?", p.BlockData.BlockId, p.TxMaps.Int64["auto_payment_id"])
	if err != nil {
		return p.ErrInfo(err)
	}*/
	return nil
}
func (p *Parser) CashRequestChatRollback() error {
	/*err := p.ExecSql("UPDATE auto_payments SET del_block_id = 0 WHERE id = ?", p.TxMaps.Int64["auto_payment_id"])
	if err != nil {
		return p.ErrInfo(err)
	}*/
	return nil
}

func (p *Parser) CashRequestChatRollbackFront() error {
	return nil
}
