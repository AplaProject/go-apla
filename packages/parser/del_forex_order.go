package parser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) DelForexOrderInit() error {

	fields := []map[string]string{{"order_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelForexOrderFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.CheckInputData(map[string]string{"order_id": "int"})
	if err != nil {
		return p.ErrInfo(err)
	}
	orderId, err := p.Single("SELECT id FROM forex_orders WHERE id  =  ? AND user_id  =  ? AND del_block_id  =  0", p.TxMaps.Int64["order_id"], p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if orderId == 0 {
		return p.ErrInfo("incorrect order_id")
	}

	// проверим, есть ли ордер для удаления
	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["order_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil || !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	return nil
}

func (p *Parser) DelForexOrder() error {
	return p.ExecSql("UPDATE forex_orders SET del_block_id = ? WHERE id = ?", p.BlockData.BlockId, p.TxMaps.Int64["order_id"])
}

func (p *Parser) DelForexOrderRollback() error {
	return p.ExecSql("UPDATE forex_orders SET del_block_id = 0 WHERE id = ?", p.TxMaps.Int64["order_id"])
}

func (p *Parser) DelForexOrderRollbackFront() error {
	return nil
}
