package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ActualizationPromisedAmountsInit() error {

	fields := []map[string]string{{"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ActualizationPromisedAmountsFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// есть ли что актуализировать
	promisedAmountId, err := p.Single("SELECT id FROM promised_amount WHERE status  =  'mining' AND user_id  =  ? AND currency_id > 1 AND del_block_id  =  0 AND del_mining_block_id  =  0 AND (cash_request_out_time > 0 AND cash_request_out_time < ? )", p.TxUserID, (p.TxTime - p.Variables.Int64["cash_request_time"])).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if promisedAmountId == 0 {
		return p.ErrInfo("incorrect promisedAmountId")
	}

	forSign := fmt.Sprintf("%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_ACTUALIZATION, "actualization", consts.LIMIT_ACTUALIZATION_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ActualizationPromisedAmounts() error {
	return p.updPromisedAmounts(p.TxUserID, false, true, 0)
}

func (p *Parser) ActualizationPromisedAmountsRollback() error {
	return p.updPromisedAmountsRollback(p.TxUserID, true)
}

func (p *Parser) ActualizationPromisedAmountsRollbackFront() error {
	return p.limitRequestsRollback("actualization")
}
