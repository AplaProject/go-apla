package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

// арбитр увеличивает время манибэка, чтобы успеть разобраться в ситуации

func (p *Parser) ChangeMoneyBackTimeInit() error {

	fields := []map[string]string{{"order_id": "int64"}, {"add_time": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeMoneyBackTimeFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"order_id": "bigint", "add_time": "bigint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.TxMaps.Int64["add_time"] == 0 || p.TxMaps.Int64["add_time"] > consts.MAX_MONEY_BACK_TIME {
		return p.ErrInfo("incorrect add_time")
	}

	// проверим, является ли арбитром для данного ордера наш юзер и не увеличивал ли он уже время
	orderId, err := p.Single("SELECT id FROM orders WHERE id  =  ? AND (arbitrator0  =  ? OR arbitrator1  =  ? OR arbitrator2  =  ? OR arbitrator3  =  ? OR arbitrator4  =  ?) AND end_time_changed  =  0 AND status  =  'refund'", p.TxMaps.Int64["order_id"], p.TxUserID, p.TxUserID, p.TxUserID, p.TxUserID, p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if orderId == 0 {
		return p.ErrInfo("incorrect order_id")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["order_id"], p.TxMap["add_time"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) ChangeMoneyBackTime() error {

	endTime, err := p.Single("SELECT end_time FROM orders WHERE id  =  ?", p.TxMaps.Int64["order_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	newEndTime := endTime + p.TxMaps.Int64["add_time"]

	err = p.selectiveLoggingAndUpd([]string{"end_time", "end_time_changed"}, []interface{}{newEndTime, 1}, "orders", []string{"id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["order_id"])})
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) ChangeMoneyBackTimeRollback() error {
	return p.selectiveRollback([]string{"end_time", "end_time_changed"}, "orders", "id="+utils.Int64ToStr(p.TxMaps.Int64["order_id"]), false)
}

func (p *Parser) ChangeMoneyBackTimeRollbackFront() error {
	return nil
}
