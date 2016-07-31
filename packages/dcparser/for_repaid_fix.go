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

func (p *Parser) ForRepaidFixInit() error {

	fields := []map[string]string{{"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ForRepaidFixFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(p.Variables.Int64["limit_for_repaid_fix"], "for_repaid_fix", p.Variables.Int64["limit_for_repaid_fix_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ForRepaidFix() error {

	// возможно больше нет mining ни по одной валюте (кроме WOC) у данного юзера
	forRepaidCurrencyIds, err := p.GetList("SELECT currency_id FROM promised_amount WHERE status  =  'mining' AND user_id  =  ? AND amount > 0 AND currency_id > 1 AND del_block_id  =  0 AND del_mining_block_id  =  0", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	var forRepaidCurrencyIdsNew []int64
	for _, currencyId := range forRepaidCurrencyIds {
		// либо сумма погашенных стала >= максимальной обещанной, т.к. в этом случае прислать этому юзеру cash_request_out будет невозможно
		maxPromisedAmount, err := p.GetMaxPromisedAmount(currencyId)
		if err != nil {
			return p.ErrInfo(err)
		}
		repaidAmount, err := p.GetRepaidAmount(currencyId, p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
		if repaidAmount < maxPromisedAmount {
			forRepaidCurrencyIdsNew = append(forRepaidCurrencyIdsNew, currencyId)
		}
	}
	if len(forRepaidCurrencyIdsNew) == 0 {
		err = p.updPromisedAmounts(p.TxUserID, false, true, 0)
		if err != nil {
			return p.ErrInfo(err)
		}
		// просроченным cash_requests ставим for_repaid_del_block_id, чтобы cash_request_out не переводил более обещанные суммы данного юзера в for_repaid из-за просроченных cash_requests
		err = p.ExecSql("UPDATE cash_requests SET for_repaid_del_block_id = ? WHERE to_user_id = ? AND time < ? AND for_repaid_del_block_id = 0", p.BlockData.BlockId, p.TxUserID, p.BlockData.Time-p.Variables.Int64["cash_request_time"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) ForRepaidFixRollback() error {
	forRepaidDelBlockId, err := p.Single("SELECT id FROM cash_requests WHERE to_user_id  =  ? AND for_repaid_del_block_id  =  ?", p.TxUserID, p.BlockData.BlockId).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if forRepaidDelBlockId > 0 {
		err = p.ExecSql("UPDATE cash_requests SET for_repaid_del_block_id = 0 WHERE to_user_id = ? AND for_repaid_del_block_id = ?", p.TxUserID, p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.updPromisedAmountsRollback(p.TxUserID, true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) ForRepaidFixRollbackFront() error {
	return p.limitRequestsRollback("for_repaid_fix")
}
