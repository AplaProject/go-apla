package dcparser

import (
	"database/sql"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) DelCfFundingInit() error {

	fields := []map[string]string{{"funding_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelCfFundingFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"funding_id": "int"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, есть ли funding для удаления
	projectId, err := p.Single("SELECT project_id FROM cf_funding WHERE id  =  ? AND user_id  =  ? AND del_block_id  =  0", p.TxMaps.Int64["funding_id"], p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if projectId == 0 {
		return p.ErrInfo("incorrect funding_id")
	}

	// проверим, на завершился ли уже проект
	projectActive, err := p.Single("SELECT id FROM cf_projects WHERE id  =  ? AND close_block_id  =  0 AND del_block_id  =  0", projectId).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if projectActive == 0 {
		return p.ErrInfo("incorrect projectId")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["funding_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) DelCfFunding() error {

	// нужно учесть набежавшие %
	var amount float64
	var time, project_id, currency_id int64
	err := p.QueryRow(p.FormatQuery("SELECT amount, time, project_id, currency_id FROM cf_funding WHERE id  =  ?"), p.TxMaps.Int64["funding_id"]).Scan(&amount, &time, &project_id, &currency_id)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}

	pointsStatus := []map[int64]string{{0: "user"}}
	pct, err := p.GetPct()
	// то, что выросло за время сбора
	profit, err := p.calcProfit_(amount, time, p.BlockData.Time, pct[currency_id], pointsStatus, [][]int64{}, []map[int64]string{}, 0, 0)
	if err != nil {
		return p.ErrInfo(err)
	}
	sumAndPct := amount + profit
	err = p.updateRecipientWallet(p.TxUserID, currency_id, sumAndPct, "cf_project_refund", project_id, "cf_project_refund", "decrypted", true)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("UPDATE cf_funding SET del_block_id = ? WHERE id = ?", p.BlockData.BlockId, p.TxMaps.Int64["funding_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) DelCfFundingRollback() error {
	err := p.ExecSql("UPDATE cf_funding SET del_block_id = 0 WHERE id = ?", p.TxMaps.Int64["funding_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	fundingData, err := p.OneRow("SELECT amount, time, currency_id FROM cf_funding WHERE id  =  ?", p.TxMaps.Int64["funding_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.generalRollback("wallets", p.TxUserID, "AND currency_id = "+fundingData["currency_id"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	// возможно были списания по кредиту
	err = p.loanPaymentsRollback(p.TxUserID, utils.StrToInt64(fundingData["currency_id"]))
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.mydctxRollback()
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelCfFundingRollbackFront() error {
	return nil
}
