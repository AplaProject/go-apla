package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"log"
	//"encoding/json"
	//"regexp"
	//"math"
	//"strings"
	//	"os"
	//	"time"
	//"strings"
	//"bytes"
	//"github.com/DayLightProject/go-daylight/packages/consts"
	//	"math"
	//	"database/sql"
	//	"bytes"
)

func (p *Parser) DelPromisedAmountInit() error {

	fields := []map[string]string{{"promised_amount_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelPromisedAmountFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"promised_amount_id": "bigint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// promised_amount должна существовать. если нет негашеных check_cash_requests, то статус promised_amount не имеет значения
	// нельзя удалить woc (currency_id=1)
	id, err := p.Single("SELECT id FROM promised_amount WHERE id  =  ? AND user_id  =  ? AND del_block_id  =  0 AND del_mining_block_id  =  0 AND currency_id > 1", p.TxMaps.Int64["promised_amount_id"], p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if id == 0 {
		return p.ErrInfo("incorrect promised_amount_id")
	}

	// У юзера не должно быть cash_requests с pending
	err = p.CheckCashRequests(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["promised_amount_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(p.Variables.Int64["limit_promised_amount"], "promised_amount", p.Variables.Int64["limit_promised_amount_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelPromisedAmount() error {
	err := p.ExecSql("UPDATE promised_amount SET del_block_id = ? WHERE id = ?", p.BlockData.BlockId, p.TxMaps.Int64["promised_amount_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	// возможно, что данный юзер имеет непогашенные cash_requests, значит новые TDC у него не растут, а просто обновляется tdc_amount_update
	newTdc, err := p.getTdc(p.TxMaps.Int64["promised_amount_id"], p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	// принудительно переводим намайненное на кошелек
	if newTdc > 0.02 {
		p.TxMaps.Money["amount"] = newTdc
		err = p.mining_(p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) DelPromisedAmountRollback() error {
	delMiningBlockId, err := p.Single("SELECT del_mining_block_id FROM promised_amount WHERE id  =  ?", p.TxMaps.Int64["promised_amount_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("delMiningBlockId %v p.BlockData.BlockId %v", delMiningBlockId, p.BlockData.BlockId)
	if delMiningBlockId == p.BlockData.BlockId {
		// выяснили, что начисление намайненного было, т.к. в методе mining() был указан del_mining_block_id. но какова сумма?
		// т.к. сумма, которая сейчас хранится в tdc_amount, равна нулю, значит предыдущую можно получить только в log_promised_amount
		logId, err := p.Single("SELECT log_id FROM promised_amount WHERE id  =  ?", p.TxMaps.Int64["promised_amount_id"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		log.Debug("logId %v", logId)
		tdcAmount, err := p.Single("SELECT tdc_and_profit FROM log_promised_amount WHERE log_id  =  ?", logId).Float64()
		if err != nil {
			return p.ErrInfo(err)
		}
		log.Debug("tdcAmount %v", tdcAmount)
		p.TxMaps.Money["amount"] = tdcAmount
		err = p.MiningRollback()
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE promised_amount SET del_mining_block_id = 0 WHERE id = ?", p.TxMaps.Int64["promised_amount_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	err = p.ExecSql("UPDATE promised_amount SET del_block_id = 0 WHERE id = ?", p.TxMaps.Int64["promised_amount_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelPromisedAmountRollbackFront() error {
	return p.limitRequestsRollback("promised_amount")
}
