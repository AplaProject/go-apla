package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) DelCfProjectInit() error {

	fields := []map[string]string{{"project_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelCfProjectFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"project_id": "int"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}
	// проверим, есть ли такой проект
	projectActive, err := p.Single("SELECT id FROM cf_projects WHERE id  =  ? AND user_id  =  ? AND close_block_id  =  0 AND del_block_id  =  0", p.TxMaps.Int64["project_id"], p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if projectActive == 0 {
		return p.ErrInfo("incorrect project_id")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["project_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) DelCfProject() error {

	err := p.ExecSql("UPDATE cf_projects SET del_block_id = ? WHERE id = ?", p.BlockData.BlockId, p.TxMaps.Int64["project_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	project_currency_id, err := p.Single("SELECT currency_id FROM cf_projects WHERE id  =  ?", p.TxMaps.Int64["project_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	pointsStatus := []map[int64]string{{0: "user"}}
	pct, err := p.GetPct()
	// проходимся по всем фундерам и возращаем на их кошельки деньги
	rows, err := p.Query(p.FormatQuery("SELECT amount, time, user_id FROM cf_funding WHERE project_id = ? AND del_block_id = 0 ORDER BY id ASC"), p.TxMaps.Int64["project_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var time, user_id int64
		var amount float64
		err = rows.Scan(&amount, &time, &user_id)
		if err != nil {
			return p.ErrInfo(err)
		}
		// то, что выросло за время сбора
		profit, err := p.calcProfit_(amount, time, p.BlockData.Time, pct[project_currency_id], pointsStatus, [][]int64{}, []map[int64]string{}, 0, 0)
		if err != nil {
			return p.ErrInfo(err)
		}
		newDCSum := amount + profit
		// возврат
		err = p.updateRecipientWallet(user_id, project_currency_id, newDCSum, "cf_project", p.TxMaps.Int64["project_id"], "cf_project", "encrypted", true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) DelCfProjectRollback() error {

	project_currency_id, err := p.Single("SELECT currency_id FROM cf_projects WHERE id  =  ?", p.TxMaps.Int64["project_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// проходимся по всем фундерам и возращаем на их кошельки деньги
	rows, err := p.Query(p.FormatQuery("SELECT user_id FROM cf_funding WHERE project_id = ? AND del_block_id = 0 ORDER BY id DESC"), p.TxMaps.Int64["project_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var user_id int64
		err = rows.Scan(&user_id)
		if err != nil {
			return p.ErrInfo(err)
		}
		// откат возврата
		err = p.generalRollback("wallets", user_id, "AND currency_id = "+utils.Int64ToStr(project_currency_id), false)
		if err != nil {
			return p.ErrInfo(err)
		}
		// возможно были списания по кредиту
		err = p.loanPaymentsRollback(user_id, project_currency_id)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	err = p.ExecSql("UPDATE cf_projects SET del_block_id = 0 WHERE id = ?", p.TxMaps.Int64["project_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelCfProjectRollbackFront() error {
	return nil
}
