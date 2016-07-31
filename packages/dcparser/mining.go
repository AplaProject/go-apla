package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) MiningInit() error {
	fields := []map[string]string{{"promised_amount_id": "int64"}, {"amount": "money"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) MiningFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"promised_amount_id": "bigint", "amount": "amount"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["promised_amount_id"], p.TxMap["amount"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// статус может быть любым кроме pending, т.к. то, что набежало в tdc_amount доступо для перевода на кошелек всегда
	num, err := p.Single("SELECT id FROM promised_amount WHERE id  =  ? AND user_id  =  ? AND status !=  'pending' AND del_block_id  =  0 AND del_mining_block_id  =  0", p.TxMaps.Int64["promised_amount_id"], p.TxMaps.Int64["user_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if num == 0 {
		return p.ErrInfo("0 promised_amount for mining")
	}

	newTdc, err := p.getTdc(p.TxMaps.Int64["promised_amount_id"], p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("newTdc", newTdc)
	log.Debug("p.TxMaps.Float64[amount]", p.TxMaps.Money["amount"])
	if newTdc < p.TxMaps.Money["amount"]+0.01 { // запас 0.01 на всяк случай
		return p.ErrInfo(fmt.Sprintf("incorrect amount %d<%f+0.01", newTdc, p.TxMaps.Money["amount"]))
	}
	if p.TxMaps.Money["amount"] < 0.02 {
		return p.ErrInfo("incorrect amount")
	}

	err = p.limitRequest(p.Variables.Int64["limit_mining"], "mining", p.Variables.Int64["limit_mining_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

/* $del_block_id указывается, когда майнинг происходит как побочный результат удаления обещанной суммы
 * */
func (p *Parser) mining_(delMiningBlockId int64) error {

	// 1 возможно нужно обновить таблицу points_status
	err := p.pointsUpdateMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	refs, err := p.getRefs(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if refs[0] > 0 {
		err = p.pointsUpdateMain(refs[0])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	if refs[1] > 0 {
		err = p.pointsUpdateMain(refs[1])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	if refs[2] > 0 {
		err = p.pointsUpdateMain(refs[2])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	data, err := p.OneRow("SELECT status, amount, currency_id, tdc_amount, tdc_amount_update, log_id FROM promised_amount WHERE id  =  ?", p.TxMaps.Int64["promised_amount_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	currencyId := utils.StrToInt64(data["currency_id"])

	// возможно, что данный юзер имеет непогашенные cash_requests, значит новые TDC у него не растут, а просто обновляется tdc_amount_update
	newTdc, err := p.getTdc(p.TxMaps.Int64["promised_amount_id"], p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// логируем текущее значение по обещанным суммам
	// tdc_and_profit - для del_promised_amount
	logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_promised_amount ( tdc_and_profit, tdc_amount, tdc_amount_update, block_id, prev_log_id ) VALUES ( ?, ?, ?, ?, ? )", "log_id", newTdc, data["tdc_amount"], data["tdc_amount_update"], p.BlockData.BlockId, data["log_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// 2 списываем сумму с promised_amount
	err = p.ExecSql("UPDATE promised_amount SET tdc_amount = ?, tdc_amount_update = ?, del_mining_block_id = ?, log_id = ? WHERE id = ?", utils.Round((newTdc-p.TxMaps.Money["amount"]), 2), p.BlockData.Time, delMiningBlockId, logId, p.TxMaps.Int64["promised_amount_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// 3 теперь начисляем DC, залогировав предыдущее значение
	err = p.updateRecipientWallet(p.TxUserID, currencyId, p.TxMaps.Money["amount"], "from_mining_id", p.TxMaps.Int64["promised_amount_id"], "", "", true)
	if err != nil {
		return p.ErrInfo(err)
	}

	// комиссия системы
	systemCommission := utils.Round(p.TxMaps.Money["amount"]*float64(float64(p.Variables.Int64["system_commission"])/100), 2)
	if systemCommission == 0 {
		log.Debug("systemCommission == 0")
		systemCommission = 0.01
	}
	if systemCommission >= p.TxMaps.Money["amount"] {
		log.Debug(`systemCommission >= p.TxMaps.Money["amount"]`)
		systemCommission = 0
	}
	log.Debug("systemCommission", systemCommission)

	// 4 теперь начисляем комиссию системе
	if systemCommission > 0 {
		err = p.updateRecipientWallet(1, currencyId, systemCommission, "system_commission", p.TxMaps.Int64["promised_amount_id"], "", "", true)
	}
	// 5 реферальные
	refData, err := p.OneRow("SELECT * FROM referral").Float64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if refs[0] > 0 {
		log.Debug("%v, %v, %v", p.TxMaps.Money["amount"], float64(refData["first"]/100), refData["first"])
		refAmount := utils.Round(p.TxMaps.Money["amount"]*float64(refData["first"]/100), 2)
		log.Debug("refs[0]", refs[0], refAmount)
		//log.Debug(p.TxMaps.Money["amount"], float64(refData["first"] / 100), refData["first"], refAmount)
		if refAmount > 0 {
			log.Debug("refAmount %v", refAmount)
			err = p.updateRecipientWallet(refs[0], currencyId, refAmount, "referral", p.TxMaps.Int64["promised_amount_id"], "", "", true)
			if err != nil {
				return p.ErrInfo(err)
			}
			// для вывода статы по рефам. табла чистится по времени
			err = p.ExecSql("INSERT INTO referral_stats ( user_id, referral, amount, currency_id, time, block_id ) VALUES ( ?, ?, ?, ?, ?, ? )", refs[0], p.TxUserID, refAmount, currencyId, p.BlockData.Time, p.BlockData.BlockId)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}

	if refs[1] > 0 {
		refAmount := utils.Round(p.TxMaps.Money["amount"]*float64(refData["second"]/100), 2)
		log.Debug("refs[1]", refs[1], refAmount)
		if refAmount > 0 {
			log.Debug("refAmount %v", refAmount)
			err = p.updateRecipientWallet(refs[1], currencyId, refAmount, "referral", p.TxMaps.Int64["promised_amount_id"], "", "", true)
			if err != nil {
				return p.ErrInfo(err)
			}
			// для вывода статы по рефам. табла чистится по времени
			err = p.ExecSql("INSERT INTO referral_stats ( user_id, referral, amount, currency_id, time, block_id ) VALUES ( ?, ?, ?, ?, ?, ? )", refs[1], refs[0], refAmount, currencyId, p.BlockData.Time, p.BlockData.BlockId)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}

	if refs[2] > 0 {
		refAmount := utils.Round(p.TxMaps.Money["amount"]*float64(refData["third"]/100), 2)
		log.Debug("refs[2]", refs[2], refAmount)
		if refAmount > 0 {
			log.Debug("refAmount %v", refAmount)
			err = p.updateRecipientWallet(refs[2], currencyId, refAmount, "referral", p.TxMaps.Int64["promised_amount_id"], "", "", true)
			if err != nil {
				return p.ErrInfo(err)
			}
			// для вывода статы по рефам. табла чистится по времени
			err = p.ExecSql("INSERT INTO referral_stats ( user_id, referral, amount, currency_id, time, block_id ) VALUES ( ?, ?, ?, ?, ?, ? )", refs[2], refs[1], refAmount, currencyId, p.BlockData.Time, p.BlockData.BlockId)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	return nil
}

func (p *Parser) Mining() error {
	return p.mining_(0)
}

func (p *Parser) MiningRollback() error {

	log.Debug("p.TxMaps.Money[amount] %v", p.TxMaps.Money["amount"])

	promisedAmountData, err := p.OneRow("SELECT * FROM promised_amount WHERE id  =  ?", p.TxMaps.Int64["promised_amount_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("promisedAmountData %v", promisedAmountData)
	refs, err := p.getRefs(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if refs[2] > 0 {
		err = p.pointsUpdateRollbackMain(refs[2])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	if refs[1] > 0 {
		err = p.pointsUpdateRollbackMain(refs[1])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	if refs[0] > 0 {
		err = p.pointsUpdateRollbackMain(refs[0])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// откатываем стату по рефам сразу по всему блоку
	err = p.ExecSql("DELETE FROM referral_stats WHERE block_id = ?", p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 5 реферальные
	var usersWalletsRollback []int64
	refData, err := p.OneRow("SELECT * FROM referral").Float64()
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("refData %v", refData)
	if refs[2] > 0 {
		log.Debug("refs[2] %v", refs[2])
		refAmount := utils.Round(p.TxMaps.Money["amount"]*float64(refData["third"]/100), 2)
		log.Debug("refAmount %v", refAmount)
		if refAmount > 0 {
			usersWalletsRollback = append(usersWalletsRollback, refs[2])
			// возможно были списания по кредиту
			err = p.loanPaymentsRollback(refs[2], utils.StrToInt64(promisedAmountData["currency_id"]))
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.generalRollback("wallets", refs[2], "AND currency_id = "+promisedAmountData["currency_id"], false)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	if refs[1] > 0 {
		log.Debug("refs[1] %v", refs[1])
		refAmount := utils.Round(p.TxMaps.Money["amount"]*float64(refData["second"]/100), 2)
		log.Debug("refAmount %v", refAmount)
		if refAmount > 0 {
			usersWalletsRollback = append(usersWalletsRollback, refs[1])
			// возможно были списания по кредиту
			err = p.loanPaymentsRollback(refs[1], utils.StrToInt64(promisedAmountData["currency_id"]))
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.generalRollback("wallets", refs[1], "AND currency_id = "+promisedAmountData["currency_id"], false)
			if err != nil {
				return p.ErrInfo(err)
			}

		}
	}
	if refs[0] > 0 {
		log.Debug("refs[0] %v", refs[0])
		refAmount := utils.Round(p.TxMaps.Money["amount"]*float64(refData["first"]/100), 2)
		log.Debug("refAmount %v", refAmount)
		if refAmount > 0 {
			// возможно были списания по кредиту
			err = p.loanPaymentsRollback(refs[0], utils.StrToInt64(promisedAmountData["currency_id"]))
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.generalRollback("wallets", refs[0], "AND currency_id = "+promisedAmountData["currency_id"], false)
			if err != nil {
				return p.ErrInfo(err)
			}
			usersWalletsRollback = append(usersWalletsRollback, refs[0])

		}
	}

	// 4 откатим комиссию системы
	systemCommission := utils.Round(p.TxMaps.Money["amount"]*float64(float64(p.Variables.Int64["system_commission"])/100), 2)
	log.Debug("systemCommission %v", systemCommission)
	if systemCommission == 0 {
		log.Debug("systemCommission == 0")
		systemCommission = 0.01
	}
	if systemCommission >= p.TxMaps.Money["amount"] {
		log.Debug(`systemCommission >= p.TxMaps.Money["amount"]`)
		systemCommission = 0
	}
	if systemCommission > 0 {
		log.Debug("systemCommission %v", systemCommission)
		log.Debug("generalRollback 1")
		// возможно были списания по кредиту
		err = p.loanPaymentsRollback(1, utils.StrToInt64(promisedAmountData["currency_id"]))
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.generalRollback("wallets", 1, "AND currency_id = "+promisedAmountData["currency_id"], false)
		if err != nil {
			return p.ErrInfo(err)
		}
		usersWalletsRollback = append(usersWalletsRollback, 1)

	}

	// возможно были списания по кредиту
	err = p.loanPaymentsRollback(p.TxUserID, utils.StrToInt64(promisedAmountData["currency_id"]))
	if err != nil {
		return p.ErrInfo(err)
	}

	// 3 откатим начисленные DC
	err = p.generalRollback("wallets", p.TxUserID, "AND currency_id = "+promisedAmountData["currency_id"], false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// данные, которые восстановим в promised_amount
	logData, err := p.OneRow("SELECT * FROM log_promised_amount WHERE log_id  =  ?", promisedAmountData["log_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	// 2 откатываем promised_amount
	err = p.ExecSql("UPDATE promised_amount SET tdc_amount = ?, tdc_amount_update = ?, log_id = ? WHERE id = ?", logData["tdc_amount"], logData["tdc_amount_update"], logData["prev_log_id"], p.TxMaps.Int64["promised_amount_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// подчищаем _log
	err = p.ExecSql("DELETE FROM log_promised_amount WHERE log_id = ?", promisedAmountData["log_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("log_promised_amount", 1)
	if err != nil {
		return p.ErrInfo(err)
	}

	// 1 возможно нужно обновить таблицу points_status
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

func (p *Parser) MiningRollbackFront() error {
	return p.limitRequestsRollback("mining")
}
