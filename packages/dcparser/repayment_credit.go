package dcparser

import (
	"database/sql"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) RepaymentCreditInit() error {

	fields := []map[string]string{{"credit_id": "int64"}, {"amount": "money"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) RepaymentCreditFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"credit_id": "bigint", "amount": "amount"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// явлется данный юзер заемщиком по данному кредиту - не проверяем.
	// т.к. нельзя запрещать кому-лбио погашать чей-либо кредит

	// не удален ли этот кредит и не погашен ли он
	currencyId, err := p.Single("SELECT currency_id FROM credits WHERE id  =  ? AND amount > 0 AND del_block_id  =  0", p.TxMaps.Int64["credit_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if currencyId == 0 {
		return p.ErrInfo("!currencyId")
	}

	// есть ли нужная сумма на кошельке
	_, err = p.checkSenderMoney(currencyId, p.TxUserID, p.TxMaps.Money["amount"], 0, 0, 0, 0, 0, 0)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["credit_id"], p.TxMap["amount"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_REPAYMENT_CREDIT, "repayment_credit", consts.REPAYMENT_CREDIT_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) RepaymentCredit() error {

	var amount float64
	var currency_id, to_user_id int64
	err := p.QueryRow(p.FormatQuery("SELECT amount, to_user_id, currency_id  FROM credits WHERE id  =  ?"), p.TxMaps.Int64["credit_id"]).Scan(&amount, &to_user_id, &currency_id)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	if p.TxMaps.Money["amount"] > amount {
		p.TxMaps.Money["amount"] = amount
	}
	// возможно нужно обновить таблицу points_status
	err = p.pointsUpdateMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	// возможно нужно обновить таблицу points_status
	err = p.pointsUpdateMain(to_user_id)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.selectiveLoggingAndUpd([]string{"amount", "tx_hash", "tx_block_id"}, []interface{}{(amount - p.TxMaps.Money["amount"]), p.TxHash, p.BlockData.BlockId}, "credits", []string{"id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["credit_id"])})
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.updateRecipientWallet(to_user_id, currency_id, p.TxMaps.Money["amount"], "loan_payment", to_user_id, "loan payment", "decrypted", false)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.updateSenderWallet(p.TxUserID, currency_id, p.TxMaps.Money["amount"], 0, "loan_payment", to_user_id, to_user_id, "loan_payment", "decrypted")
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) RepaymentCreditRollback() error {
	creditData, err := p.OneRow("SELECT to_user_id, currency_id FROM credits WHERE id  =  ?", p.TxMaps.Int64["credit_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.generalRollback("wallets", p.TxUserID, "AND currency_id = "+utils.Int64ToStr(creditData["currency_id"]), false)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.generalRollback("wallets", creditData["to_user_id"], "AND currency_id = "+utils.Int64ToStr(creditData["currency_id"]), false)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.selectiveRollback([]string{"amount", "tx_hash", "tx_block_id"}, "credits", "id="+utils.Int64ToStr(p.TxMaps.Int64["credit_id"]), false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// возможно нужно обновить таблицу points_status
	err = p.pointsUpdateMain(creditData["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// возможно нужно обновить таблицу points_status
	err = p.pointsUpdateMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) RepaymentCreditRollbackFront() error {
	return p.limitRequestsRollback("repayment_credit")
}
