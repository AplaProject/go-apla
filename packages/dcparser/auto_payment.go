package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) AutoPaymentInit() error {
	fields := []map[string]string{{"auto_payment_id": "int64"}, {"sign": "bytes"}}

	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxMaps.String["comment"] == "null" {
		p.TxMaps.String["comment"] = ""
		p.TxMap["comment"] = []byte("")
	}
	return nil
}

func (p *Parser) AutoPaymentFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"auto_payment_id": "bigint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	var txTime int64
	if p.BlockData != nil {
		txTime = p.BlockData.Time
	} else {
		txTime = utils.Time() - 30 // просто на всякий случай небольшой запас
	}

	var autoPaymentData map[string]string
	if p.BlockData == nil || p.BlockData.BlockId > 289840 {
		// проверим, существует ли такой платеж и прошло ли нужное кол-во времени с последнего платежа
		autoPaymentData, err = p.OneRow(`
			SELECT id, commission, currency_id, amount
			FROM auto_payments
			WHERE id = ? AND sender	= ?	AND last_payment_time < ? - period AND del_block_id = 0
			LIMIT 1
			`, p.TxMaps.Int64["auto_payment_id"], p.TxUserID, txTime).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		if len(autoPaymentData["id"]) == 0 {
			return p.ErrInfo("autoPaymentData==0")
		}
	}
	commission := utils.StrToFloat64(autoPaymentData["commission"])
	amount := utils.StrToFloat64(autoPaymentData["amount"])
	currencyId := utils.StrToInt64(autoPaymentData["currency_id"])

	nodeCommission, err := p.getMyNodeCommission(currencyId, p.TxUserID, amount)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, удовлетворяет ли нас комиссия, которую предлагает юзер
	if commission < nodeCommission {
		return p.ErrInfo(fmt.Sprintf("commission %v<%v", commission, nodeCommission))
	}

	// нодовский ключ
	nodePublicKey, err := p.GetNodePublicKey(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(nodePublicKey) == 0 {
		return p.ErrInfo("incorrect user_id")
	}
	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["auto_payment_id"])
	CheckSignResult, err := utils.CheckSign([][]byte{nodePublicKey}, forSign, p.TxMap["sign"], true)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	_, err = p.checkSenderMoney(currencyId, p.TxUserID, amount, commission, 0, 0, 0, 0, 0)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) AutoPayment() error {

	autoPaymentData, err := p.OneRow(`
			SELECT commission, amount, currency_id, recipient
			FROM auto_payments
			WHERE id = ?
			LIMIT 1
			`, p.TxMaps.Int64["auto_payment_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	commission := utils.StrToFloat64(autoPaymentData["commission"])
	amount := utils.StrToFloat64(autoPaymentData["amount"])
	currencyId := utils.StrToInt64(autoPaymentData["currency_id"])
	recipient := utils.StrToInt64(autoPaymentData["recipient"])

	// 1 возможно нужно обновить таблицу points_status
	err = p.pointsUpdateMain(p.BlockData.UserId)
	if err != nil {
		return p.ErrInfo(err)
	}
	// 2
	err = p.pointsUpdateMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	// 3
	err = p.pointsUpdateMain(recipient)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 4 обновим сумму на кошельке отправителя, залогировав предыдущее значение
	err = p.updateSenderWallet(p.TxUserID, currencyId, amount, commission, "from_user", recipient, recipient, string(utils.BinToHex(p.TxMap["comment"])), "encrypted")
	if err != nil {
		return p.ErrInfo(err)
	}

	log.Debug("AutoPayment updateRecipientWallet")
	// 5 обновим сумму на кошельке получателю
	err = p.updateRecipientWallet(recipient, currencyId, amount, "from_user", p.TxUserID, p.TxMaps.String["comment"], "encrypted", true)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 6 теперь начисляем комиссию майнеру, который этот блок сгенерил
	if commission >= 0.01 {
		err = p.updateRecipientWallet(p.BlockData.UserId, currencyId, commission, "node_commission", p.BlockData.BlockId, "", "encrypted", true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// отмечаем данную транзакцию в буфере как отработанную и ставим в очередь на удаление
	err = p.ExecSql("UPDATE wallets_buffer SET del_block_id = ? WHERE hex(hash) = ?", p.BlockData.BlockId, p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.selectiveLoggingAndUpd([]string{"last_payment_time"}, []interface{}{p.BlockData.Time}, "auto_payments", []string{"id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["auto_payment_id"])})
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) AutoPaymentRollback() error {

	autoPaymentData, err := p.OneRow(`
			SELECT commission, currency_id, recipient
			FROM auto_payments
			WHERE id = ?
			LIMIT 1
			`, p.TxMaps.Int64["auto_payment_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	commission := utils.StrToFloat64(autoPaymentData["commission"])
	currencyId := utils.StrToInt64(autoPaymentData["currency_id"])
	recipient := utils.StrToInt64(autoPaymentData["recipient"])


	// 6 комиссия нода-генератора блока
	if commission >= 0.01 {
		err = p.generalRollback("wallets", p.BlockData.UserId, "AND currency_id = "+autoPaymentData["currency_id"], false)
		if err != nil {
			return p.ErrInfo(err)
		}
		// возможно были списания по кредиту нода-генератора
		err = p.loanPaymentsRollback(p.BlockData.UserId, currencyId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// 5 обновим сумму на кошельке получателю
	// возможно были списания по кредиту
	err = p.loanPaymentsRollback(recipient, currencyId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.generalRollback("wallets", recipient, "AND currency_id = "+autoPaymentData["currency_id"], false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 4 обновим сумму на кошельке отправителя
	err = p.generalRollback("wallets", p.TxUserID, "AND currency_id = "+autoPaymentData["currency_id"], false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 3
	err = p.pointsUpdateRollbackMain(recipient)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 2
	err = p.pointsUpdateRollbackMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	// 1
	err = p.pointsUpdateRollbackMain(p.BlockData.UserId)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.mydctxRollback()
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.selectiveRollback([]string{"last_payment_time"}, "auto_payments", "id="+utils.Int64ToStr(p.TxMaps.Int64["auto_payment_id"]), false)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) AutoPaymentRollbackFront() error {
	return nil

}
