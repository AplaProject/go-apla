package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) DelCreditInit() error {

	fields := []map[string]string{{"credit_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelCreditFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"credit_id": "bigint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// явлется данный юзер кредитором
	id, err := p.Single("SELECT id FROM credits WHERE id  =  ? AND to_user_id  =  ? AND del_block_id  =  0", p.TxMaps.Int64["credit_id"], p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if id == 0 {
		return p.ErrInfo("not a creditor")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["credit_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) DelCredit() error {
	return p.ExecSql("UPDATE credits SET del_block_id = ? WHERE id = ?", p.BlockData.BlockId, p.TxMaps.Int64["credit_id"])
}

func (p *Parser) DelCreditRollback() error {
	return p.ExecSql("UPDATE credits SET del_block_id = 0 WHERE id = ?", p.TxMaps.Int64["credit_id"])
}

func (p *Parser) DelCreditRollbackFront() error {
	return nil
}
