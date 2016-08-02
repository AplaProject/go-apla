package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) NewForexOrderInit() error {

	fields := []map[string]string{{"sell_currency_id": "int64"}, {"sell_rate": "float64"}, {"amount": "money"}, {"buy_currency_id": "int64"}, {"commission": "money"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	/*
		sell_currency_id Что продается
		sell_rate По какому курсу к buy_currency_id
		amount сколько продается
		buy_currency_id Какая валюта нужна
		commission Сколько готовы отдать комиссию ноду-генератору
	*/
	return nil
}

func (p *Parser) NewForexOrderFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"sell_currency_id": "int", "sell_rate": "sell_rate", "amount": "amount", "buy_currency_id": "int", "commission": "amount"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.TxMaps.Int64["sell_currency_id"] == p.TxMaps.Int64["buy_currency_id"] {
		return p.ErrInfo("sell_currency_id == buy_currency_id")
	}
	if p.TxMaps.Float64["sell_rate"] == 0 {
		return p.ErrInfo("sell_rate=0")
	}
	if p.TxMaps.Money["amount"] == 0 {
		return p.ErrInfo("amount=0")
	}
	if p.TxMaps.Money["amount"]*p.TxMaps.Float64["sell_rate"] < 0.01 {
		return p.ErrInfo("amount * sell_rate < 0.01")
	}

	checkCurrency, err := p.CheckCurrency(p.TxMaps.Int64["sell_currency_id"])
	if err != nil || !checkCurrency {
		return p.ErrInfo("!sell_currency_id")
	}
	checkCurrency, err = p.CheckCurrency(p.TxMaps.Int64["buy_currency_id"])
	if err != nil || !checkCurrency {
		return p.ErrInfo("!buy_currency_id")
	}

	p.TxMaps.Int64["currency_id"] = p.TxMaps.Int64["sell_currency_id"]
	nodeCommission, err := p.getMyNodeCommission(p.TxMaps.Int64["currency_id"], p.TxUserID, p.TxMaps.Money["amount"])

	// проверим, удовлетворяет ли нас комиссия, которую предлагает юзер
	if p.TxMaps.Money["commission"] < nodeCommission {
		return p.ErrInfo("commission")
	}
	// есть ли нужная сумма на кошельке
	p.TxMaps.Int64["from_user_id"] = p.TxMaps.Int64["user_id"]
	for i := 0; i < 5; i++ {
		p.TxMaps.Float64["arbitrator"+utils.IntToStr(i)+"_commission"] = 0
	}
	_, err = p.checkSenderMoney(p.TxMaps.Int64["currency_id"], p.TxMaps.Int64["from_user_id"], p.TxMaps.Money["amount"], p.TxMaps.Money["commission"], p.TxMaps.Float64["arbitrator0_commission"], p.TxMaps.Float64["arbitrator1_commission"], p.TxMaps.Float64["arbitrator2_commission"], p.TxMaps.Float64["arbitrator3_commission"], p.TxMaps.Float64["arbitrator4_commission"])
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["sell_currency_id"], p.TxMap["sell_rate"], p.TxMap["amount"], p.TxMap["buy_currency_id"], p.TxMap["commission"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.checkSpamMoney(p.TxMaps.Int64["sell_currency_id"], p.TxMaps.Money["amount"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewForexOrder() error {

	// нужно отметить в log_time_money_orders, что тр-ия прошла в блок
	err := p.ExecSql("UPDATE log_time_money_orders SET del_block_id = ? WHERE hex(tx_hash) = ?", p.BlockData.BlockId, p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}

	// логируем, чтобы можно было делать откат. Важен только сам ID
	mainId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_forex_orders_main ( block_id ) VALUES ( ? )", "id", p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}

	// обратный курс. нужен для поиска по ордерам
	reverseRate := 1 / p.TxMaps.Float64["sell_rate"]

	// Сколько хотим потратить
	totalSellAmount := p.TxMaps.Money["amount"]

	// прежде всего начислим комиссию ноду-генератору
	if p.TxMaps.Money["commission"] >= 0.01 {
		// возможно нужно обновить таблицу points_status
		err = p.pointsUpdateMain(p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.updateSenderWallet(p.TxUserID, p.TxMaps.Int64["sell_currency_id"], p.TxMaps.Money["commission"], 0, "from_user", p.TxUserID, p.BlockData.WalletId, "node_commission", "decrypted")
		if err != nil {
			return p.ErrInfo(err)
		}
		// возможно нужно обновить таблицу points_status
		err = p.pointsUpdateMain(p.BlockData.WalletId)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.updateRecipientWallet(p.BlockData.WalletId, p.TxMaps.Int64["sell_currency_id"], p.TxMaps.Money["commission"], "node_commission", p.BlockData.BlockId, "", "encrypted", true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// берем из БД только те ордеры, которые удовлетворяют нашим требованиям
	rows, err := p.Query(p.FormatQuery("SELECT id, sell_rate, amount, user_id, sell_currency_id, buy_currency_id  FROM forex_orders WHERE buy_currency_id = ? AND sell_rate >= ? AND sell_currency_id = ? AND del_block_id = 0 AND empty_block_id = 0"), p.TxMaps.Int64["sell_currency_id"], reverseRate, p.TxMaps.Int64["buy_currency_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	defer rows.Close()
	for rows.Next() {
		var rowId, rowUserId, rowSellCurrencyId, rowBuyCurrencyId int64
		var rowSellRate, rowAmount float64
		err = rows.Scan(&rowId, &rowSellRate, &rowAmount, &rowUserId, &rowSellCurrencyId, &rowBuyCurrencyId)
		if err != nil {
			return p.ErrInfo(err)
		}
		log.Debug("rowId", rowId, "rowUserId", rowUserId, "rowSellCurrencyId", rowSellCurrencyId, "rowBuyCurrencyId", rowBuyCurrencyId, "rowSellRate", rowSellRate, "rowAmount", rowAmount)
		// сколько мы готовы купить по курсу владельца данного ордера
		readyToBuy := totalSellAmount * rowSellRate

		// сколько продавец данного ордера продал валюты
		var sellerSellAmount float64
		if readyToBuy >= rowAmount {
			sellerSellAmount = rowAmount // ордер будет закрыт, а мы продолжим искать новые
		} else {
			sellerSellAmount = readyToBuy // данный ордер удовлетворяет наш запрос целиком
		}
		if rowAmount-sellerSellAmount < 0.01 { // ордер опустошили
			err = p.ExecSql("UPDATE forex_orders SET amount = 0, empty_block_id = ? WHERE id = ?", p.BlockData.BlockId, rowId)
			if err != nil {
				return p.ErrInfo(err)
			}
		} else {
			// вычитаем забранную сумму из ордера
			err = p.ExecSql("UPDATE forex_orders SET amount = round(amount - ?, 2) WHERE id = ?", sellerSellAmount, rowId)
			if err != nil {
				return p.ErrInfo(err)
			}
		}

		// логируем данную операцию
		err = p.ExecSql("INSERT INTO log_forex_orders ( main_id, order_id, amount, to_user_id, block_id ) VALUES ( ?, ?, round(?, 2), ?, ? )", mainId, rowId, sellerSellAmount, p.TxUserID, p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}

		// === Продавец валюты (тот, чей ордер обработали) ===

		// сколько продавец получил с продажи суммы $seller_sell_amount по его курсу
		sellerBuyAmount := sellerSellAmount * (1 / rowSellRate)

		// ===  списываем валюту, которую продавец продал (U) ===

		// возможно нужно обновить таблицу points_status
		err = p.pointsUpdateMain(rowUserId)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.updateSenderWallet(rowUserId, rowSellCurrencyId, sellerSellAmount, 0, "from_user", rowUserId, p.TxUserID, "order # "+utils.Int64ToStr(rowId), "decrypted")
		if err != nil {
			return p.ErrInfo(err)
		}

		// === начисляем валюту, которую продавец получил (R) ===

		// возможно нужно обновить таблицу points_status
		err = p.pointsUpdateMain(rowUserId)
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.updateRecipientWallet(rowUserId, rowBuyCurrencyId, sellerBuyAmount, "from_user", p.TxUserID, "order # "+utils.Int64ToStr(rowId), "decrypted", true)
		if err != nil {
			return p.ErrInfo(err)
		}

		// ===  Покупатель валюты (наш юзер)

		// списываем валюту, которую мы продали (R)

		// возможно нужно обновить таблицу points_status
		err = p.pointsUpdateMain(p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.updateSenderWallet(p.TxUserID, rowBuyCurrencyId, sellerBuyAmount, 0, "from_user", p.TxUserID, rowUserId, "order # "+utils.Int64ToStr(rowId), "decrypted")
		if err != nil {
			return p.ErrInfo(err)
		}

		// начисляем валюту, которую мы получили (U)

		// возможно нужно обновить таблицу points_status
		err = p.pointsUpdateMain(p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.updateRecipientWallet(p.TxUserID, rowSellCurrencyId, sellerSellAmount, "from_user", rowUserId, "order # "+utils.Int64ToStr(rowId), "decrypted", true)
		if err != nil {
			return p.ErrInfo(err)
		}

		// вычитаем с нашего баланса сумму, которую потратили на данный ордер
		totalSellAmount -= sellerBuyAmount
		if totalSellAmount < 0.01 {
			break // проход по ордерам прекращаем, т.к. наш запрос удовлетворен
		}
	}

	// если после прохода по всем имеющимся ордерам мы не потратили все средства, то создаем свой ордер
	if totalSellAmount >= 0.01 {
		orderId, err := p.ExecSqlGetLastInsertId("INSERT INTO forex_orders ( user_id, sell_currency_id, sell_rate, amount, buy_currency_id, commission ) VALUES ( ?, ?, ?, round(?, 2), ?, ? )", "id", p.TxUserID, p.TxMaps.Int64["sell_currency_id"], p.TxMaps.Float64["sell_rate"], totalSellAmount, p.TxMaps.Int64["buy_currency_id"], p.TxMaps.Money["commission"])
		if err != nil {
			return p.ErrInfo(err)
		}

		// логируем данную операцию. amount не указывается, т.к. при откате будет просто удалена запись из forex_orders
		err = p.ExecSql("INSERT INTO log_forex_orders ( main_id, order_id, new, block_id ) VALUES ( ?, ?, 1, ? )", mainId, orderId, p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) NewForexOrderRollback() error {

	// нужно отметить в log_time_money_orders, что тр-ия НЕ прошла в блок
	err := p.ExecSql("UPDATE log_time_money_orders SET del_block_id = 0 WHERE hex(tx_hash) = ?", p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}

	// откат всегда идет по последней записи в log_forex_orders_main
	mainId, err := p.Single("SELECT id FROM log_forex_orders_main ORDER BY id DESC").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// проходимся по всем ордерам, которые затронула данная тр-ия
	rows, err := p.Query(p.FormatQuery(`
				SELECT   log_forex_orders.amount,
							  log_forex_orders.id,
							  empty_block_id,
							  order_id,
							  user_id,
							  to_user_id,
							  new,
							  commission,
							  buy_currency_id,
							  sell_rate,
							  sell_currency_id
				FROM log_forex_orders
				LEFT JOIN forex_orders ON log_forex_orders.order_id = forex_orders.id
				WHERE main_id = ?
				ORDER BY log_forex_orders.id DESC
	`), mainId)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var rowId, rowEmptyBlockId, rowOrderId, rowUserId, rowToUserId, rowNew, rowBuyCurrencyId, rowSellCurrencyId int64
		var rowAmount, rowCommission, rowSellRate float64
		err = rows.Scan(&rowAmount, &rowId, &rowEmptyBlockId, &rowOrderId, &rowUserId, &rowToUserId, &rowNew, &rowCommission, &rowBuyCurrencyId, &rowSellRate, &rowSellCurrencyId)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("DELETE FROM log_forex_orders WHERE id = ?", rowId)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.rollbackAI("log_forex_orders", 1)

		// если это создание нового ордера, то просто удалим его
		if rowNew > 0 {
			err = p.ExecSql("DELETE FROM forex_orders WHERE id = ?", rowOrderId)
			if err != nil {
				return p.ErrInfo(err)
			}
			p.rollbackAI("forex_orders", 1)
			// берем следующий ордер
			// никаких движений средств не произошло, откатывать кошельки не нужно
		} else {
			addSql := ""
			if rowEmptyBlockId == p.BlockData.BlockId {
				addSql = ", empty_block_id = 0"
			}

			// вернем amount ордеру
			err = p.ExecSql("UPDATE forex_orders SET amount = round(amount + ?, 2) "+addSql+" WHERE id = ?", rowAmount, rowOrderId)
			if err != nil {
				return p.ErrInfo(err)
			}

			// откатываем покупателя (наш юзер)

			// возможно нужно обновить таблицу points_status
			err = p.pointsUpdateRollbackMain(rowToUserId)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.generalRollback("wallets", rowToUserId, "AND currency_id ="+utils.Int64ToStr(rowSellCurrencyId), false)
			if err != nil {
				return p.ErrInfo(err)
			}

			// возможно были списания по кредиту
			err = p.loanPaymentsRollback(rowToUserId, rowSellCurrencyId)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.generalRollback("wallets", rowToUserId, "AND currency_id ="+utils.Int64ToStr(rowBuyCurrencyId), false)
			if err != nil {
				return p.ErrInfo(err)
			}
			// откатим продавца

			// возможно нужно обновить таблицу points_status
			err = p.pointsUpdateRollbackMain(rowUserId)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.generalRollback("wallets", rowUserId, "AND currency_id ="+utils.Int64ToStr(rowBuyCurrencyId), false)
			if err != nil {
				return p.ErrInfo(err)
			}

			// возможно были списания по кредиту
			err = p.loanPaymentsRollback(rowUserId, rowBuyCurrencyId)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.generalRollback("wallets", rowUserId, "AND currency_id ="+utils.Int64ToStr(rowSellCurrencyId), false)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	// откатим комиссию ноду-генератору
	if p.TxMaps.Money["commission"] >= 0.01 {

		// возможно нужно обновить таблицу points_status
		err = p.pointsUpdateRollbackMain(p.BlockData.WalletId)
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.generalRollback("wallets", p.BlockData.WalletId, "AND currency_id ="+utils.Int64ToStr(p.TxMaps.Int64["sell_currency_id"]), false)
		if err != nil {
			return p.ErrInfo(err)
		}

		// возможно были списания по кредиту
		err = p.loanPaymentsRollback(p.BlockData.WalletId, p.TxMaps.Int64["sell_currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}

		// возможно нужно обновить таблицу points_status
		err = p.pointsUpdateRollbackMain(p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.generalRollback("wallets", p.TxUserID, "AND currency_id ="+utils.Int64ToStr(p.TxMaps.Int64["sell_currency_id"]), false)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// и напоследок удалим запись, из-за которой откат был инициирован
	err = p.ExecSql("DELETE FROM log_forex_orders_main WHERE id = ?", mainId)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.rollbackAI("log_forex_orders_main", 1)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.mydctxRollback()
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewForexOrderRollbackFront() error {

	err := p.limitRequestsMoneyOrdersRollback()
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil

}
