package parser

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"math"
)

func (p *Parser) SendDcInit() error {
	var fields []map[string]string
	if p.BlockData != nil && p.BlockData.BlockId <= consts.ARBITRATION_BLOCK_START {
		fields = []map[string]string{{"to_user_id": "int64"}, {"currency_id": "int64"}, {"amount": "float64"}, {"commission": "float64"}, {"comment": "string"}, {"sign": "bytes"}}
	} else {
		fields = []map[string]string{{"to_user_id": "int64"}, {"currency_id": "int64"}, {"amount": "float64"}, {"commission": "float64"}, {"arbitrator0": "int64"}, {"arbitrator1": "int64"}, {"arbitrator2": "int64"}, {"arbitrator3": "int64"}, {"arbitrator4": "int64"}, {"arbitrator0_commission": "float64"}, {"arbitrator1_commission": "float64"}, {"arbitrator2_commission": "float64"}, {"arbitrator3_commission": "float64"}, {"arbitrator4_commission": "float64"}, {"comment": "string"}, {"sign": "bytes"}}
	}

	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.Bytes["hash_hex"] = utils.BinToHex(p.TxMaps.Bytes["hash"])
	p.TxMaps.Int64["from_user_id"] = p.TxMaps.Int64["user_id"]
	p.TxMap["hash_hex"] = utils.BinToHex(p.TxMap["hash"])
	p.TxMap["from_user_id"] = p.TxMap["user_id"]
	if p.TxMaps.String["comment"] == "null" {
		p.TxMaps.String["comment"] = ""
		p.TxMap["comment"] = []byte("")
	}
	return nil
}

func (p *Parser) SendDcFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"from_user_id": "bigint", "to_user_id": "bigint", "currency_id": "bigint", "amount": "amount", "commission": "amount", "comment": "comment"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.TxMaps.Float64["amount"] < 0.01 { // 0.01 - минимальная сумма
		return p.ErrInfo("amount")
	}

	// проверим, существует ли такая валюта в таблиуе DC-валют
	checkCurrency, err := p.CheckCurrency(p.TxMaps.Int64["currency_id"])
	if !checkCurrency {
		// если нет, то проверяем список CF-валют
		checkCurrency, err := p.CheckCurrencyCF(p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		if !checkCurrency {
			return p.ErrInfo("currency_id")
		}
	}

	nodeCommission, err := p.getMyNodeCommission(p.TxMaps.Int64["currency_id"], p.TxUserID, p.TxMaps.Float64["amount"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, удовлетворяет ли нас комиссия, которую предлагает юзер
	if p.TxMaps.Float64["commission"] < nodeCommission {
		return p.ErrInfo(fmt.Sprintf("commission %v<%v", p.TxMaps.Float64["commission"], nodeCommission))
	}

	if p.BlockData != nil && p.BlockData.BlockId <= consts.ARBITRATION_BLOCK_START {
		for i := 0; i < 5; i++ {
			p.TxMaps.Float64["arbitrator"+utils.IntToStr(i)+"_commission"] = 0 // для check_sender_money
		}
		forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["to_user_id"], p.TxMap["amount"], p.TxMap["commission"], utils.BinToHex(p.TxMap["comment"]), p.TxMap["currency_id"])
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
		if err != nil {
			return p.ErrInfo(err)
		}
		if !CheckSignResult {
			return p.ErrInfo("incorrect sign")
		}
	} else {
		dupArray := make(map[int64]int64)
		for i := 0; i < 5; i++ {
			arbitrator__commission := "arbitrator" + utils.IntToStr(i) + "_commission"
			arbitrator_ := "arbitrator" + utils.IntToStr(i)
			if !utils.CheckInputData(p.TxMap[arbitrator__commission], "amount") {
				return p.ErrInfo("arbitrator_commission")
			}
			if !utils.CheckInputData(p.TxMap[arbitrator_], "bigint") {
				return p.ErrInfo("arbitrator")
			}
			// если указал ID арбитра, то должна быть комиссия для него
			if p.TxMaps.Int64[arbitrator_] > 0 && p.TxMaps.Float64[arbitrator__commission] < 0.01 {
				return p.ErrInfo("arbitrator_commission")
			}

			// на всяк случай не даем арбитрам возможность быть арбитрами самим себе
			if p.TxMaps.Int64[arbitrator_] == p.TxUserID {
				return p.ErrInfo("arbitrator = user_id")
			}

			if p.TxMaps.Int64[arbitrator_] > 0 {
				dupArray[utils.BytesToInt64(p.TxMap[arbitrator_])]++
				if dupArray[utils.BytesToInt64(p.TxMap[arbitrator_])] > 1 {
					return p.ErrInfo("doubles")
				}
			}
			if p.TxMaps.Int64[arbitrator_] > 0 {

				arbitrator := p.TxMap[arbitrator_]
				// проверим, является ли арбитром указанный user_id
				arbitratorConditionsJson, err := p.Single("SELECT conditions FROM arbitrator_conditions WHERE user_id  =  ?", utils.BytesToInt64(arbitrator)).Bytes()
				if err != nil {
					return p.ErrInfo(err)
				}
				arbitratorConditionsMap := make(map[string][5]string)
				err = json.Unmarshal(arbitratorConditionsJson, &arbitratorConditionsMap)
				// арбитр к этому моменту мог передумать и убрать свои условия, уйдя из арбитров для новых сделок поставив [0] что вызовет тут ошибку
				if err != nil {
					log.Debug("arbitratorConditionsJson", arbitratorConditionsJson)
					return p.ErrInfo(err)
				}
				// проверим, работает ли выбранный арбитр с валютой данной сделки
				var checkCurrency int64
				if p.TxMaps.Int64["currency_id"] > 1000 {
					checkCurrency = 1000
				} else {
					checkCurrency = p.TxMaps.Int64["currency_id"]
				}
				if len(arbitratorConditionsMap[utils.Int64ToStr(checkCurrency)]) == 0 {
					return p.ErrInfo("len(arbitratorConditionsMap[checkCurrency]) == 0")
				}
				// указан ли этот арбитр в списке доверенных у продавца
				sellerArbitrator, err := p.Single("SELECT user_id FROM arbitration_trust_list WHERE user_id  =  ? AND arbitrator_user_id  =  ?", p.TxMaps.Int64["to_user_id"], utils.BytesToInt64(arbitrator)).Int64()
				if err != nil {
					return p.ErrInfo(err)
				}
				if sellerArbitrator == 0 {
					return p.ErrInfo("sellerArbitrator == 0")
				}
				// указан ли этот арбитр в списке доверенных у покупателя
				buyerArbitrator, err := p.Single("SELECT user_id FROM arbitration_trust_list WHERE user_id  =  ? AND arbitrator_user_id  =  ?", p.TxMaps.Int64["from_user_id"], utils.BytesToInt64(arbitrator)).Int64()
				if err != nil {
					return p.ErrInfo(err)
				}
				if buyerArbitrator == 0 {
					return p.ErrInfo("buyerArbitrator == 0")
				}
				// согласен ли продавец на манибек
				arbitrationDaysRefund, err := p.Single("SELECT arbitration_days_refund FROM users WHERE user_id  =  ?", p.TxMaps.Int64["to_user_id"]).Int64()
				if err != nil {
					return p.ErrInfo(err)
				}
				if arbitrationDaysRefund == 0 {
					return p.ErrInfo("buyerArbitrator == 0")
				}
				// готов ли арбитр рассматривать такую сумму сделки
				currencyIdStr := utils.Int64ToStr(p.TxMaps.Int64["currency_id"])
				if p.TxMaps.Float64["amount"] < utils.StrToFloat64(arbitratorConditionsMap[currencyIdStr][0]) ||
					(p.TxMaps.Float64["amount"] > utils.StrToFloat64(arbitratorConditionsMap[currencyIdStr][1]) && utils.StrToFloat64(arbitratorConditionsMap[currencyIdStr][1]) > 0) {
					return p.ErrInfo("amount")
				}
				// мин. комиссия, на которую согласен арбитр
				minArbitratorCommission := utils.StrToFloat64(arbitratorConditionsMap[currencyIdStr][4]) / 100 * p.TxMaps.Float64["amount"]
				if minArbitratorCommission > utils.StrToFloat64(arbitratorConditionsMap[currencyIdStr][3]) && utils.StrToFloat64(arbitratorConditionsMap[currencyIdStr][3]) > 0 {
					minArbitratorCommission = utils.StrToFloat64(arbitratorConditionsMap[currencyIdStr][3])
				}
				if minArbitratorCommission < utils.StrToFloat64(arbitratorConditionsMap[currencyIdStr][2]) {
					minArbitratorCommission = utils.StrToFloat64(arbitratorConditionsMap[currencyIdStr][2])
				}
				if utils.BytesToFloat64(p.TxMap[arbitrator__commission]) < minArbitratorCommission {
					return p.ErrInfo(" < minArbitratorCommission")
				}
			}
		}

		forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["to_user_id"], p.TxMap["amount"], p.TxMap["commission"], p.TxMap["arbitrator0"], p.TxMap["arbitrator1"], p.TxMap["arbitrator2"], p.TxMap["arbitrator3"], p.TxMap["arbitrator4"], p.TxMap["arbitrator0_commission"], p.TxMap["arbitrator1_commission"], p.TxMap["arbitrator2_commission"], p.TxMap["arbitrator3_commission"], p.TxMap["arbitrator4_commission"], utils.BinToHex(p.TxMap["comment"]), p.TxMap["currency_id"])
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
		if err != nil {
			return p.ErrInfo(err)
		}
		if !CheckSignResult {
			return p.ErrInfo("incorrect sign")
		}
	}

	/*
			wallets_buffer сделан не для защиты от двойной траты, а для того, чтобы нода, которая генерит блок не записала
			двойное списание в свой блок, который будет отправлен другим нодам и будет ими отвергнут.
		   Для тр-ий типа new_forex_order используется простой запрет на запись в блок тр-ии new_forex_order+new_forex_order или
		   new_forex_order+send_dc и пр. Сейчас это сделано в виде запрета более 1 тр-ии от 1 юзера в 1 блоке.
		   защита от двойного списания на основе даннных из блока, полученного из сети заключается в постепенной обработке
		   тр-ий путем проверки front_ и занесения данных в БД (ParseDataFull).
	*/
	amountAndCommission, err := p.checkSenderMoney(p.TxMaps.Int64["currency_id"], p.TxMaps.Int64["from_user_id"], p.TxMaps.Float64["amount"], p.TxMaps.Float64["commission"], p.TxMaps.Float64["arbitrator0_commission"], p.TxMaps.Float64["arbitrator1_commission"], p.TxMaps.Float64["arbitrator2_commission"], p.TxMaps.Float64["arbitrator3_commission"], p.TxMaps.Float64["arbitrator4_commission"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// существует ли юзер-получатель
	err = p.CheckUser(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.checkSpamMoney(p.TxMaps.Int64["currency_id"], p.TxMaps.Float64["amount"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// вычитаем из wallets_buffer
	// amount_and_commission взято из check_sender_money()
	err = p.updateWalletsBuffer(amountAndCommission, p.TxMaps.Int64["currency_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) SendDc() error {
	// нужно отметить в log_time_money_orders, что тр-ия прошла в блок
	err := p.ExecSql("UPDATE log_time_money_orders SET del_block_id = ? WHERE hex(tx_hash) = ?", p.BlockData.BlockId, p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 1 возможно нужно обновить таблицу points_status
	err = p.pointsUpdateMain(p.BlockData.WalletId)
	if err != nil {
		return p.ErrInfo(err)
	}
	// 2
	err = p.pointsUpdateMain(p.TxMaps.Int64["from_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	// 3
	err = p.pointsUpdateMain(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	commission := p.TxMaps.Float64["commission"]

	var arbitration_days_refund int64
	var seller_hold_back_pct float64
	var arbitrators bool
	if p.BlockData.BlockId > consts.ARBITRATION_BLOCK_START {
		// на какой период манибека согласен продавец
		err := p.QueryRow(p.FormatQuery("SELECT arbitration_days_refund, seller_hold_back_pct FROM users WHERE user_id  =  ?"), p.TxMaps.Int64["to_user_id"]).Scan(&arbitration_days_refund, &seller_hold_back_pct)
		if err != nil && err != sql.ErrNoRows {
			return p.ErrInfo(err)
		}
		if arbitration_days_refund > 0 {
			for i := 0; i < 5; i++ {
				iStr := utils.IntToStr(i)
				if p.TxMaps.Int64["arbitrator"+iStr] > 0 && p.TxMaps.Float64["arbitrator"+iStr+"_commission"] >= 0.01 {
					// нужно учесть комиссию арбитра
					commission += p.TxMaps.Float64["arbitrator"+iStr+"_commission"]
					arbitrators = true
				}
			}
		}
	}
	// 4 обновим сумму на кошельке отправителя, залогировав предыдущее значение
	err = p.updateSenderWallet(p.TxMaps.Int64["from_user_id"], p.TxMaps.Int64["currency_id"], p.TxMaps.Float64["amount"], commission, "from_user", p.TxMaps.Int64["to_user_id"], p.TxMaps.Int64["to_user_id"], string(utils.BinToHex(p.TxMap["comment"])), "encrypted")
	if err != nil {
		return p.ErrInfo(err)
	}

	log.Debug("SendDC updateRecipientWallet")
	// 5 обновим сумму на кошельке получателю
	err = p.updateRecipientWallet(p.TxMaps.Int64["to_user_id"], p.TxMaps.Int64["currency_id"], p.TxMaps.Float64["amount"], "from_user", p.TxMaps.Int64["from_user_id"], p.TxMaps.String["comment"], "encrypted", true)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 6 теперь начисляем комиссию майнеру, который этот блок сгенерил
	if p.TxMaps.Float64["commission"] >= 0.01 {
		err = p.updateRecipientWallet(p.BlockData.WalletId, p.TxMaps.Int64["currency_id"], p.TxMaps.Float64["commission"], "node_commission", p.BlockData.BlockId, "", "encrypted", true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	log.Debug("p.BlockData.BlockId", p.BlockData.BlockId)
	log.Debug("consts.ARBITRATION_BLOCK_START", consts.ARBITRATION_BLOCK_START)
	if p.BlockData.BlockId > consts.ARBITRATION_BLOCK_START {

		// если продавец не согласен на арбитраж, то $arbitration_days_refund будет равно 0

		log.Debug("arbitration_days_refund", arbitration_days_refund)
		log.Debug("arbitrators", arbitrators)
		if arbitration_days_refund > 0 && arbitrators {
			holdBackAmount := math.Floor(utils.Round(p.TxMaps.Float64["amount"]*(seller_hold_back_pct/100), 3)*100) / 100
			if holdBackAmount < 0.01 {
				holdBackAmount = 0.01
			}
			orderId, err := p.ExecSqlGetLastInsertId("INSERT INTO orders ( time, buyer, seller, arbitrator0, arbitrator1, arbitrator2, arbitrator3, arbitrator4, amount, hold_back_amount, currency_id, block_id, end_time ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )", "id", p.BlockData.Time, p.TxMaps.Int64["from_user_id"], p.TxMaps.Int64["to_user_id"], p.TxMaps.Int64["arbitrator0"], p.TxMaps.Int64["arbitrator1"], p.TxMaps.Int64["arbitrator2"], p.TxMaps.Int64["arbitrator3"], p.TxMaps.Int64["arbitrator4"], p.TxMaps.Float64["amount"], holdBackAmount, p.TxMaps.Int64["currency_id"], p.BlockData.BlockId, (p.BlockData.Time + arbitration_days_refund*86400))
			if err != nil {
				return p.ErrInfo(err)
			}
			// начисляем комиссию арбитрам
			for i := 0; i < 5; i++ {
				iStr := utils.IntToStr(i)
				log.Debug("arbitrator_commission "+iStr, p.TxMaps.Float64["arbitrator"+iStr+"_commission"])
				if p.TxMaps.Int64["arbitrator"+iStr] > 0 && p.TxMaps.Float64["arbitrator"+iStr+"_commission"] >= 0.01 {
					log.Debug("updateRecipientWallet")
					err = p.updateRecipientWallet(p.TxMaps.Int64["arbitrator"+iStr], p.TxMaps.Int64["currency_id"], p.TxMaps.Float64["arbitrator"+iStr+"_commission"], "arbitrator_commission", orderId, "", "encrypted", true)
					if err != nil {
						return p.ErrInfo(err)
					}
				}
			}
		}
	}
	// отмечаем данную транзакцию в буфере как отработанную и ставим в очередь на удаление
	err = p.ExecSql("UPDATE wallets_buffer SET del_block_id = ? WHERE hex(hash) = ?", p.BlockData.BlockId, p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) SendDcRollback() error {
	// нужно отметить в log_time_money_orders, что тр-ия НЕ прошла в блок
	err := p.ExecSql("UPDATE log_time_money_orders SET del_block_id = 0 WHERE hex(tx_hash) = ?", p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 6 комиссия нода-генератора блока
	if p.TxMaps.Float64["commission"] >= 0.01 {
		err = p.generalRollback("wallets", p.BlockData.WalletId, "AND currency_id = "+utils.Int64ToStr(p.TxMaps.Int64["currency_id"]), false)
		if err != nil {
			return p.ErrInfo(err)
		}
		// возможно были списания по кредиту нода-генератора
		err = p.loanPaymentsRollback(p.BlockData.WalletId, p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// 5 обновим сумму на кошельке получателю
	// возможно были списания по кредиту
	err = p.loanPaymentsRollback(p.TxMaps.Int64["to_user_id"], p.TxMaps.Int64["currency_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.generalRollback("wallets", p.TxMaps.Int64["to_user_id"], "AND currency_id = "+utils.Int64ToStr(p.TxMaps.Int64["currency_id"]), false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 4 обновим сумму на кошельке отправителя
	err = p.generalRollback("wallets", p.TxMaps.Int64["from_user_id"], "AND currency_id = "+utils.Int64ToStr(p.TxMaps.Int64["currency_id"]), false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 3
	err = p.pointsUpdateRollbackMain(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// 2
	err = p.pointsUpdateRollbackMain(p.TxMaps.Int64["from_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	// 1
	err = p.pointsUpdateRollbackMain(p.BlockData.WalletId)
	if err != nil {
		return p.ErrInfo(err)
	}

	// отменяем чистку буфера
	err = p.ExecSql("UPDATE wallets_buffer SET del_block_id = 0 WHERE hex(hash) = ?", p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.BlockData.BlockId > consts.ARBITRATION_BLOCK_START {

		// на какой период манибека согласен продавец
		arbitrationDaysRefund, err := p.Single("SELECT arbitration_days_refund FROM users WHERE user_id  =  ?", p.TxMaps.Int64["to_user_id"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if arbitrationDaysRefund > 0 {
			delOrder := false
			for i := 0; i < 5; i++ {
				iStr := utils.IntToStr(i)
				if p.TxMaps.Int64["arbitrator"+iStr] > 0 && p.TxMaps.Float64["arbitrator"+iStr+"_commission"] >= 0.01 {
					err = p.generalRollback("wallets", p.TxMaps.Int64["arbitrator"+iStr], "AND currency_id = "+utils.Int64ToStr(p.TxMaps.Int64["currency_id"]), false)
					if err != nil {
						return p.ErrInfo(err)
					}
					if !delOrder {
						err = p.ExecSql("DELETE FROM orders WHERE block_id = ? AND buyer = ? AND seller = ? AND currency_id = ? AND amount = ?", p.BlockData.BlockId, p.TxMaps.Int64["from_user_id"], p.TxMaps.Int64["to_user_id"], p.TxMaps.Int64["currency_id"], p.TxMaps.Float64["amount"])
						if err != nil {
							return p.ErrInfo(err)
						}
						err = p.rollbackAI("orders", 1)
						if err != nil {
							return p.ErrInfo(err)
						}
						delOrder = true
					}
					// возможно были списания по кредиту арбитра
					err = p.loanPaymentsRollback(p.TxMaps.Int64["arbitrator"+iStr], p.TxMaps.Int64["currency_id"])
					if err != nil {
						return p.ErrInfo(err)
					}
				}
			}
		}
	}

	err = p.mydctxRollback()
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) SendDcRollbackFront() error {
	err := p.ExecSql("DELETE FROM dlt_wallets_buffer WHERE hex(hash) = ?", p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.limitRequestsMoneyOrdersRollback()
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil

}
