package dcparser

import (
	"database/sql"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"math"
	"time"
)

func (p *Parser) MoneyBackInit() error {

	fields := []map[string]string{{"order_id": "int64"}, {"amount": "money"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) MoneyBackFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"order_id": "bigint", "amount": "amount"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	var txTime int64
	if p.BlockData != nil {
		txTime = p.BlockData.Time
	} else {
		txTime = time.Now().Unix() - 30 // просто на всякий случай небольшой запас
	}

	// проверим корректность ордера. тр-ия может быть как от продавца, так и от арбитра
	orderId, err := p.Single(`
				SELECT id
				FROM orders
				WHERE id = ? AND
							 status = 'refund' AND
							 (
									(
										 (arbitrator0 = ? OR arbitrator1 = ? OR arbitrator2 = ? OR arbitrator3 = ? OR arbitrator4 = ?) AND
										 refund = 0 AND
										  (amount - voluntary_refund - ?) >= 0 AND
										 end_time >= ?
									)
									OR (
										 seller = ? AND
										 voluntary_refund = 0 AND
										  (amount - refund - ?) >= 0 AND
										 end_time >= ?
									)
							)
				LIMIT 1
	`, p.TxMaps.Int64["order_id"], p.TxUserID, p.TxUserID, p.TxUserID, p.TxUserID, p.TxUserID, string(p.TxMap["amount"]), txTime, p.TxUserID, string(p.TxMap["amount"]), txTime).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if orderId == 0 {
		return p.ErrInfo("orderId==0")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["order_id"], p.TxMap["amount"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) MoneyBack() error {

	data, err := p.OneRow("SELECT buyer, seller, currency_id FROM orders WHERE id  =  ?", p.TxMaps.Int64["order_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	buyerUserId := data["buyer"]
	sellerUserId := data["seller"]
	p.TxMaps.Int64["currency_id"] = data["currency_id"]

	// возможно нужно обновить таблицу points_status
	err = p.pointsUpdateMain(sellerUserId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.pointsUpdateMain(sellerUserId)
	if err != nil {
		return p.ErrInfo(err)
	}

	// если на счету продавца есть не вся сумма, то на остаток будет создан кредит
	pct, err := p.GetPct()

	// получим сумму на кошельке юзера + %
	var wallet_amount float64
	var last_update int64
	err = p.QueryRow(p.FormatQuery("SELECT amount, last_update FROM wallets WHERE user_id = ? AND currency_id = ?"), sellerUserId, p.TxMaps.Int64["currency_id"]).Scan(&wallet_amount, &last_update)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	profit, err := p.calcProfit_(wallet_amount, last_update, p.BlockData.Time, pct[p.TxMaps.Int64["currency_id"]], []map[int64]string{{0: "user"}}, [][]int64{}, []map[int64]string{}, 0, 0)
	if err != nil {
		return p.ErrInfo(err)
	}
	totalAmount := wallet_amount + profit

	// учтем все свежие cash_requests, которые висят со статусом pending
	cashRequestsAmount, err := p.Single("SELECT sum(amount) FROM cash_requests WHERE from_user_id  =  ? AND currency_id  =  ? AND status  =  'pending' AND time > ?", sellerUserId, p.TxMaps.Int64["currency_id"], (p.BlockData.Time - p.Variables.Int64["cash_request_time"])).Float64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// учитываются все fx-ордеры
	forexOrdersAmount, err := p.Single("SELECT sum(amount) FROM forex_orders WHERE user_id  =  ? AND sell_currency_id  =  ? AND del_block_id  =  0", sellerUserId, p.TxMaps.Int64["currency_id"]).Float64()
	if err != nil {
		return p.ErrInfo(err)
	}
	// НЕ учитываем все текущие суммы холдбека, т.к. кроме этой суммы у продавца может ничего и не быть
	all := math.Floor((totalAmount-cashRequestsAmount-forexOrdersAmount)*100) / 100
	var amount float64
	if all >= p.TxMaps.Money["amount"] {
		amount = p.TxMaps.Money["amount"]
	} else {
		amount = all
		creditAmount := p.TxMaps.Money["amount"] - amount
		err = p.ExecSql("INSERT INTO credits ( time, amount, from_user_id, to_user_id, currency_id, pct, tx_hash, tx_block_id ) VALUES ( ?, ?, ?, ?, ?, 100, [hex], ? )", p.BlockData.Time, creditAmount, sellerUserId, buyerUserId, p.TxMaps.Int64["currency_id"], p.TxHash, p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// если на счету продавца еще что-то есть, то делаем перевод покупателю
	if amount >= 0.01 {
		err = p.updateSenderWallet(sellerUserId, p.TxMaps.Int64["currency_id"], amount, 0, "money_back", p.TxMaps.Int64["order_id"], buyerUserId, "money_back", "decrypted")
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.updateRecipientWallet(buyerUserId, p.TxMaps.Int64["currency_id"], p.TxMaps.Money["amount"], "money_back", p.TxMaps.Int64["order_id"], "money_back", "encrypted", true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	if p.TxUserID == sellerUserId {
		// отмечаем, какую сумму вернул продавец, чтобы арбитр её учел
		err := p.selectiveLoggingAndUpd([]string{"voluntary_refund"}, []interface{}{p.TxMaps.Money["amount"]}, "orders", []string{"id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["order_id"])})
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		// отмечаем, какую сумму вернул арбитр, чтобы продавец её учел при доп. манибеке
		err := p.selectiveLoggingAndUpd([]string{"refund", "refund_arbitrator_id", "arbitrator_refund_time"}, []interface{}{p.TxMaps.Money["amount"], p.TxUserID, p.BlockData.Time}, "orders", []string{"id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["order_id"])})
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) MoneyBackRollback() error {

	data, err := p.OneRow("SELECT buyer, seller, currency_id FROM orders WHERE id  =  ?", p.TxMaps.Int64["order_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	buyerUserId := data["buyer"]
	sellerUserId := data["seller"]
	p.TxMaps.Int64["currency_id"] = data["currency_id"]

	// возможно нужно обновить таблицу points_status
	err = p.pointsUpdateRollbackMain(buyerUserId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.pointsUpdateRollbackMain(sellerUserId)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxUserID == sellerUserId {
		// отмечаем, какую сумму вернул продавец, чтобы арбитр её учел
		err := p.selectiveRollback([]string{"voluntary_refund"}, "orders", "id="+utils.Int64ToStr(p.TxMaps.Int64["order_id"]), false)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		// отмечаем, какую сумму вернул арбитр, чтобы продавец её учел при доп. манибеке
		err := p.selectiveRollback([]string{"refund", "refund_arbitrator_id", "arbitrator_refund_time"}, "orders", "id="+utils.Int64ToStr(p.TxMaps.Int64["order_id"]), false)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	var rollbackWallet bool
	// если был создан кредит, значит у продавца не хватило денег на счету
	creditAmount, err := p.Single("SELECT amount FROM credits WHERE tx_block_id  =  ? AND hex(tx_hash) = ?", p.BlockData.BlockId, p.TxHash).Float64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if creditAmount > 0 {
		// если сумма кредита меньше суммы возврата, то что-то было списано со счета продавца
		if p.TxMaps.Money["amount"] < creditAmount {
			rollbackWallet = true
		}
		err = p.ExecSql("DELETE FROM credits WHERE tx_block_id = ? AND hex(tx_hash) = ?", p.BlockData.BlockId, p.TxHash)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.rollbackAI("credits", 1)
	} else {
		rollbackWallet = true
	}
	if rollbackWallet {
		err = p.generalRollback("wallets", buyerUserId, "AND currency_id = "+utils.Int64ToStr(p.TxMaps.Int64["currency_id"]), false)
		if err != nil {
			return p.ErrInfo(err)
		}
		// возможно были списания по кредиту
		err = p.loanPaymentsRollback(buyerUserId, p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.generalRollback("wallets", sellerUserId, "AND currency_id = "+utils.Int64ToStr(p.TxMaps.Int64["currency_id"]), false)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	err = p.mydctxRollback()
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) MoneyBackRollbackFront() error {
	return nil
}
