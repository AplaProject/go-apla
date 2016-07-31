package dcparser

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) NewMaxPromisedAmountsInit() error {
	var err error
	var fields []string
	fields = []string{"new_max_promised_amounts", "sign"}
	p.TxMap, err = p.GetTxMap(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}
func (p *Parser) NewMaxPromisedAmountsFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	nodePublicKey, err := p.GetNodePublicKey(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(nodePublicKey) == 0 {
		return p.ErrInfo("incorrect user_id")
	}

	allMaxAmounts := utils.GetAllMaxPromisedAmount()
	if err != nil {
		return p.ErrInfo(err)
	}

	totalCountCurrencies, err := p.GetCountCurrencies()
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, верно ли указаны ID валют
	currencyList := make(map[string]int64)
	err = json.Unmarshal(p.TxMap["new_max_promised_amounts"], &currencyList)
	if err != nil {
		return p.ErrInfo(err)
	}
	currencyIdsSql := ""
	var countCurrency int64
	for currencyId, amount := range currencyList {
		if !utils.CheckInputData(currencyId, "int") {
			return p.ErrInfo("currencyId")
		}
		currencyIdsSql += currencyId + ","
		countCurrency++
		if !utils.InSliceInt64(amount, allMaxAmounts) {
			return p.ErrInfo("incorrect amount")
		}
	}
	currencyIdsSql = currencyIdsSql[0 : len(currencyIdsSql)-1]
	if countCurrency == 0 {
		return p.ErrInfo("countCurrency")
	}
	count, err := p.Single("SELECT count(id) FROM currency WHERE id IN (" + currencyIdsSql + ")").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if count != countCurrency {
		return p.ErrInfo("count != countCurrency")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["new_max_promised_amounts"])
	CheckSignResult, err := utils.CheckSign([][]byte{nodePublicKey}, forSign, p.TxMap["sign"], true)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// проверим, прошло ли 2 недели с момента последнего обновления
	pctTime, err := p.Single("SELECT max(time) FROM max_promised_amounts").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxTime-pctTime <= p.Variables.Int64["new_max_promised_amount"] {
		return p.ErrInfo("14 day error")
	}

	// берем все голоса
	maxPromisedAmountVotes := make(map[int64][]map[int64]int64)
	rows, err := p.Query("SELECT currency_id, amount, count(user_id) as votes FROM votes_max_promised_amount GROUP BY currency_id, amount ORDER BY currency_id, amount ASC")
	if err != nil {
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var currency_id, amount, votes int64
		err = rows.Scan(&currency_id, &amount, &votes)
		if err != nil {
			return p.ErrInfo(err)
		}
		maxPromisedAmountVotes[currency_id] = append(maxPromisedAmountVotes[currency_id], map[int64]int64{amount: votes})
		//fmt.Println("currency_id", currency_id)
	}

	NewMaxPromisedAmountsVotes := make(map[string]int64)
	for currencyId, amountsAndVotes := range maxPromisedAmountVotes {
		NewMaxPromisedAmountsVotes[utils.Int64ToStr(currencyId)] = utils.GetMaxVote(amountsAndVotes, 0, totalCountCurrencies, 10)
	}

	jsonData, err := json.Marshal(NewMaxPromisedAmountsVotes)
	if err != nil {
		return p.ErrInfo(err)
	}
	if string(p.TxMap["new_max_promised_amounts"]) != string(jsonData) {
		return p.ErrInfo("p.TxMap[new_max_promised_amounts] != jsonData " + string(p.TxMap["new_max_promised_amounts"]) + "!=" + string(jsonData))
	}

	return nil
}

func (p *Parser) NewMaxPromisedAmounts() error {

	newMaxPromisedAmounts := make(map[string]int64)
	err := json.Unmarshal(p.TxMap["new_max_promised_amounts"], &newMaxPromisedAmounts)
	if err != nil {
		return p.ErrInfo(err)
	}

	for currencyId, amount := range newMaxPromisedAmounts {
		err = p.ExecSql("INSERT INTO max_promised_amounts ( time, currency_id, amount, block_id ) VALUES ( ?, ?, ?, ? )", p.BlockData.Time, currencyId, amount, p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) NewMaxPromisedAmountsRollback() error {
	affect, err := p.ExecSqlGetAffect("DELETE FROM max_promised_amounts WHERE block_id = ?", p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("max_promised_amounts", affect)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewMaxPromisedAmountsRollbackFront() error {
	return nil
}
