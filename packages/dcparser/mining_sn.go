package dcparser

import (
	//"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
)

func (p *Parser) MiningSnInit() error {
	fields := []map[string]string{{"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) MiningSnFront() error {

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

	// юзер должен иметь статус sn_user
	status, err := p.Single("SELECT status FROM users WHERE user_id = ?", p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if status != "sn_user" {
		return p.ErrInfo(`status != "sn_user"`)
	}

	restrictedPA, err := p.OneRow(`SELECT * from promised_amount_restricted WHERE currency_id = 72 AND user_id = ?`, p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(restrictedPA) == 0 {
		return p.ErrInfo("promised_amount_restricted == 0")
	}
	pct, err := p.GetPct()
	if err != nil {
		return p.ErrInfo(err)
	}
	startTime := utils.StrToInt64(restrictedPA["last_update"])
	var txTime int64
	if p.BlockData != nil { // тр-ия пришла в блоке
		txTime = p.BlockData.Time
	} else {
		txTime = utils.Time() - 30 // просто на всякий случай небольшой запас
	}
	profit, err := p.calcProfit_(utils.StrToFloat64(restrictedPA["amount"]), startTime, txTime, pct[72], []map[int64]string{{0: "user"}}, [][]int64{}, []map[int64]string{}, 0, 0)
	if err != nil {
		return p.ErrInfo(err)
	}
	fmt.Println("profit", profit)

	newDcAmount := utils.StrToFloat64(restrictedPA["dc_amount"]) + profit;
	if newDcAmount > 30 {
		newDcAmount = 30;
		profit = 30 - utils.StrToFloat64(restrictedPA["dc_amount"]);
	}
	profit = utils.Round(profit, 2);

	// можно получить нахаляву максимум 30 dUSD
	if profit < 0.01 || utils.StrToFloat64(restrictedPA["dc_amount"]) > 29.99 {
		return p.ErrInfo(fmt.Sprintf("incorrect amount %v %v", profit, restrictedPA["dc_amount"] ))
	}

	err = p.limitRequest(p.Variables.Int64["limit_mining"], "mining", p.Variables.Int64["limit_mining_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) MiningSn() error {
	// 1 возможно нужно обновить таблицу points_status
	err := p.pointsUpdateMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	restrictedPA, err := p.OneRow(`SELECT * from promised_amount_restricted WHERE currency_id = 72 AND user_id = ?`, p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	startTime := utils.StrToInt64(restrictedPA["last_update"])
	var txTime int64
	if p.BlockData != nil { // тр-ия пришла в блоке
		txTime = p.BlockData.Time
	} else {
		txTime = utils.Time() - 30 // просто на всякий случай небольшой запас
	}
	pct, err := p.GetPct()
	if err != nil {
		return p.ErrInfo(err)
	}
	profit, err := p.calcProfit_(utils.StrToFloat64(restrictedPA["amount"]), startTime, txTime, pct[72], []map[int64]string{{0: "user"}}, [][]int64{}, []map[int64]string{}, 0, 0)
	if err != nil {
		return p.ErrInfo(err)
	}

	newDcAmount := utils.StrToFloat64(restrictedPA["dc_amount"]) + profit;
	if newDcAmount > 30 {
		newDcAmount = 30;
		profit = 30 - utils.StrToFloat64(restrictedPA["dc_amount"]);
	}
	profit = utils.Round(profit, 2);

	err = p.selectiveLoggingAndUpd([]string{"dc_amount", "last_update"}, []interface{}{newDcAmount, p.BlockData.Time}, "promised_amount_restricted", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.updateRecipientWallet(p.TxUserID, 72, profit, "from_mining_id", 1, "", "", true)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) MiningSnRollback() error {

	// возможно были списания по кредиту
	err := p.loanPaymentsRollback(p.TxUserID, 72)
	if err != nil {
		return p.ErrInfo(err)
	}

	// откатим начисленные DC
	err = p.generalRollback("wallets", p.TxUserID, "AND currency_id = 72", false)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.selectiveRollback([]string{"dc_amount", "last_update"}, "promised_amount_restricted", "user_id="+utils.Int64ToStr(p.TxUserID), false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// возможно нужно обновить таблицу points_status
	err = p.pointsUpdateRollbackMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.mydctxRollback()
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) MiningSnRollbackFront() error {
	return p.limitRequestsRollback("mining")
}
