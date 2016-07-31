package dcparser

import (
	"database/sql"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) CfSendDcInit() error {
	fields := []map[string]string{{"project_id": "int64"}, {"amount": "money"}, {"commission": "money"}, {"comment": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CfSendDcFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"project_id": "int", "amount": "amount", "commission": "amount", "comment": "comment"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxMaps.Money["amount"] < 0.01 { // 0.01 - минимальная сумма
		return p.ErrInfo("amount<0.01")
	}

	// не закончился ли сбор средств
	projectCurrencyId, err := p.Single("SELECT currency_id FROM cf_projects WHERE id  =  ? AND close_block_id  =  0 AND del_block_id  =  0", p.TxMaps.Int64["project_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if projectCurrencyId == 0 {
		return p.ErrInfo("projectCurrencyId==0")
	}

	nodeCommission, err := p.getMyNodeCommission(projectCurrencyId, p.TxUserID, p.TxMaps.Money["amount"])
	if err != nil {
		return p.ErrInfo(err)
	}
	// проверим, удовлетворяет ли нас комиссия, которую предлагает юзер
	if p.TxMaps.Money["commission"] < nodeCommission {
		return p.ErrInfo("incorrect commission")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["project_id"], p.TxMap["amount"], p.TxMap["commission"], utils.BinToHex(p.TxMap["comment"]))
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// Для защиты от несовместимости тр-ий cf_send_dc, new_forex_order, send_dc,cash_requests не могут быть в одном блоке (clear_incompatible_tx()). А cf_send_dc, new_forex_order,cash_requests могут быть только в единичном кол-ве в одном блоке от одного юзера.

	// есть ли нужная сумма в кошельке
	_, err = p.checkSenderMoney(projectCurrencyId, p.TxUserID, p.TxMaps.Money["amount"], p.TxMaps.Money["commission"], 0, 0, 0, 0, 0)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.limitRequest(consts.LIMIT_CF_SEND_DC, "cf_send_dc", consts.LIMIT_CF_SEND_DC_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CfSendDc() error {
	var cf_projects_amount float64
	var cf_projects_project_currency_name string
	var cf_projects_end_time, cf_projects_currency_id, cf_projects_user_id int64
	err := p.QueryRow(p.FormatQuery("SELECT amount, end_time, currency_id, project_currency_name, user_id FROM cf_projects WHERE id  =  ?"), p.TxMaps.Int64["project_id"]).Scan(&cf_projects_amount, &cf_projects_end_time, &cf_projects_currency_id, &cf_projects_project_currency_name, &cf_projects_user_id)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	// возможно нужно обновить таблицу points_status
	err = p.pointsUpdateMain(p.BlockData.UserId)
	// возможно нужно обновить таблицу points_status
	err = p.pointsUpdateMain(p.TxUserID)

	// начисляем комиссию майнеру, который этот блок сгенерил
	if p.TxMaps.Money["commission"] >= 0.01 {
		err = p.updateRecipientWallet(p.BlockData.UserId, cf_projects_currency_id, p.TxMaps.Money["commission"], "node_commission", p.BlockData.BlockId, "", "encrypted", true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	fundingId, err := p.ExecSqlGetLastInsertId("INSERT INTO cf_funding ( project_id, user_id, amount, currency_id, time, block_id ) VALUES ( ?, ?, ?, ?, ?, ? )", "id", p.TxMaps.Int64["project_id"], p.TxUserID, p.TxMaps.Money["amount"], cf_projects_currency_id, p.BlockData.Time, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновим сумму на кошельке отправителя, залогировав предыдущее значение
	err = p.updateSenderWallet(p.TxUserID, cf_projects_currency_id, p.TxMaps.Money["amount"], p.TxMaps.Money["commission"], "cf_project", p.TxUserID, fundingId, string(utils.BinToHex(p.TxMap["comment"])), "encrypted")
	if err != nil {
		return p.ErrInfo(err)
	}

	// если время сбора средств закончилось
	if cf_projects_end_time < p.BlockData.Time {

		// закрываем проект
		err = p.ExecSql("UPDATE cf_projects SET close_block_id = ? WHERE id = ?", p.BlockData.BlockId, p.TxMaps.Int64["project_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		sum, err := p.Single("SELECT sum(amount) FROM cf_funding WHERE project_id  =  ? AND del_block_id  =  0", p.TxMaps.Int64["project_id"]).Float64()
		if err != nil {
			return p.ErrInfo(err)
		}
		pointsStatus := []map[int64]string{{0: "user"}}

		pct, err := p.GetPct()
		if err != nil {
			return p.ErrInfo(err)
		}
		// нужная сумма набрана
		if sum >= cf_projects_amount {

			// запишем в таблицу CF-валют новую валюту и получим ID
			projectCurrencyId, err := p.ExecSqlGetLastInsertId("INSERT INTO cf_currency ( name, project_id ) VALUES ( ?, ? )", "id", cf_projects_project_currency_name, p.TxMaps.Int64["project_id"])
			if err != nil {
				return p.ErrInfo(err)
			}

			// начисляем общую сумму на кошелек автора проекта
			// а также, начисляем бэкерам валюту проекта
			// нужно учесть набежавшие %
			rows, err := p.Query(p.FormatQuery("SELECT amount, time, user_id FROM cf_funding WHERE project_id = ? AND del_block_id = 0 ORDER BY id ASC"), p.TxMaps.Int64["project_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
			defer rows.Close()
			var sumAndPctAuthor float64
			for rows.Next() {
				var cf_funding_amount float64
				var cf_funding_time, cf_funding_user_id int64
				err = rows.Scan(&cf_funding_amount, &cf_funding_time, &cf_funding_user_id)
				if err != nil {
					return p.ErrInfo(err)
				}
				// то, что выросло за время сбора
				profit, err := p.calcProfit_(cf_funding_amount, cf_funding_time, p.BlockData.Time, pct[cf_projects_currency_id], pointsStatus, [][]int64{}, []map[int64]string{}, 0, 0)
				if err != nil {
					return p.ErrInfo(err)
				}
				amountAndPct := cf_funding_amount + profit

				// автору проекта обычные DC
				sumAndPctAuthor += amountAndPct

				// бэкерам - валюта проекта
				err = p.updateRecipientWallet(cf_funding_user_id, projectCurrencyId, amountAndPct, "cf_project", p.TxMaps.Int64["project_id"], "cf_project", "encrypted", true)
				if err != nil {
					return p.ErrInfo(err)
				}
			}
			// автору - DC
			err = p.updateRecipientWallet(cf_projects_user_id, cf_projects_currency_id, sumAndPctAuthor, "cf_project", p.TxMaps.Int64["project_id"], "cf_project", "encrypted", true)
			if err != nil {
				return p.ErrInfo(err)
			}
		} else { // нужная сумма не набрана

			// проходимся по всем фундерам и возращаем на их кошельки деньги
			rows, err := p.Query(p.FormatQuery("SELECT amount, time, user_id FROM cf_funding WHERE project_id = ? AND del_block_id = 0 ORDER BY id ASC"), p.TxMaps.Int64["project_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
			defer rows.Close()
			for rows.Next() {
				var cf_funding_amount float64
				var cf_funding_time, cf_funding_user_id int64
				err = rows.Scan(&cf_funding_amount, &cf_funding_time, &cf_funding_user_id)
				if err != nil {
					return p.ErrInfo(err)
				}
				// то, что выросло за время сбора
				profit, err := p.calcProfit_(cf_funding_amount, cf_funding_time, p.BlockData.Time, pct[cf_projects_currency_id], pointsStatus, [][]int64{}, []map[int64]string{}, 0, 0)
				if err != nil {
					return p.ErrInfo(err)
				}
				newDCSum := cf_funding_amount + profit
				// возврат
				err = p.updateRecipientWallet(cf_funding_user_id, cf_projects_currency_id, newDCSum, "cf_project", p.TxMaps.Int64["project_id"], "cf_project", "encrypted", true)
				if err != nil {
					return p.ErrInfo(err)
				}

			}
		}
	}
	return nil
}

func (p *Parser) CfSendDcRollback() error {

	var cf_projects_amount float64
	var cf_projects_project_currency_name string
	var cf_projects_end_time, cf_projects_currency_id, cf_projects_user_id int64
	err := p.QueryRow(p.FormatQuery("SELECT amount, end_time, currency_id, project_currency_name, user_id FROM cf_projects WHERE id  =  ?"), p.TxMaps.Int64["project_id"]).Scan(&cf_projects_amount, &cf_projects_end_time, &cf_projects_currency_id, &cf_projects_project_currency_name, &cf_projects_user_id)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}

	// если время сбора средств закончилось
	if cf_projects_end_time <= p.BlockData.Time {
		sum, err := p.Single("SELECT sum(amount) FROM cf_funding WHERE project_id  =  ? AND del_block_id  =  0", p.TxMaps.Int64["project_id"]).Float64()
		if err != nil {
			return p.ErrInfo(err)
		}
		// нужная сумма набрана
		if sum >= cf_projects_amount {

			// откатываем начисление общей суммы на кошелек автора проекта
			err = p.generalRollback("wallets", cf_projects_user_id, "AND currency_id = "+utils.Int64ToStr(cf_projects_currency_id), false)
			if err != nil {
				return p.ErrInfo(err)
			}

			// возможно были списания по кредиту
			err = p.loanPaymentsRollback(cf_projects_user_id, cf_projects_currency_id)
			if err != nil {
				return p.ErrInfo(err)
			}

			// узнаем ID валюты, которая была создана
			projectCurrencyId, err := p.Single("SELECT id FROM cf_currency WHERE name  =  ?", cf_projects_project_currency_name).Int64()
			if err != nil {
				return p.ErrInfo(err)
			}

			// проходимся по всем фундерам и забираем у них начисленную валюту проекта
			rows, err := p.Query(p.FormatQuery("SELECT user_id FROM cf_funding WHERE project_id = ? AND del_block_id = 0 ORDER BY id DESC"), p.TxMaps.Int64["project_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
			defer rows.Close()
			for rows.Next() {
				var cf_funding_user_id int64
				err = rows.Scan(&cf_funding_user_id)
				if err != nil {
					return p.ErrInfo(err)
				}
				// откат возврата
				err = p.generalRollback("wallets", cf_funding_user_id, "AND currency_id = "+utils.Int64ToStr(projectCurrencyId), false)
				if err != nil {
					return p.ErrInfo(err)
				}

				// возможно были списания по кредиту
				err = p.loanPaymentsRollback(cf_funding_user_id, projectCurrencyId)
				if err != nil {
					return p.ErrInfo(err)
				}
			}

			// Удаляем созданную валюту
			err = p.ExecSql("DELETE FROM cf_currency WHERE name = ?", cf_projects_project_currency_name)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.rollbackAI("cf_currency", 1)
			if err != nil {
				return p.ErrInfo(err)
			}
		} else { // нужная сумма не набрана

			// проходимся по всем фундерам и возращаем на их кошельки деньги
			rows, err := p.Query(p.FormatQuery("SELECT user_id FROM cf_funding WHERE project_id = ? AND del_block_id = 0 ORDER BY id DESC"), p.TxMaps.Int64["project_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
			defer rows.Close()
			for rows.Next() {
				var cf_funding_user_id int64
				err = rows.Scan(&cf_funding_user_id)
				if err != nil {
					return p.ErrInfo(err)
				}
				// откат возврата
				err = p.generalRollback("wallets", cf_funding_user_id, "AND currency_id = "+utils.Int64ToStr(cf_projects_currency_id), false)
				if err != nil {
					return p.ErrInfo(err)
				}

				// возможно были списания по кредиту
				err = p.loanPaymentsRollback(cf_funding_user_id, cf_projects_currency_id)
				if err != nil {
					return p.ErrInfo(err)
				}
			}
		}

		// откатываем закрытие проекта
		err = p.ExecSql("UPDATE cf_projects SET close_block_id = 0 WHERE id = ?", p.TxMaps.Int64["project_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	err = p.ExecSql("DELETE FROM cf_funding WHERE block_id = ? AND user_id = ? AND project_id = ?", p.BlockData.BlockId, p.TxUserID, p.TxMaps.Int64["project_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("cf_funding", 1)
	if err != nil {
		return p.ErrInfo(err)
	}

	// откат списания средств с кошелька фундера
	err = p.generalRollback("wallets", p.TxUserID, "AND currency_id = "+utils.Int64ToStr(cf_projects_currency_id), false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// откат комиссии
	if p.TxMaps.Money["commission"] >= 0.01 {
		err = p.generalRollback("wallets", p.BlockData.UserId, "AND currency_id = "+utils.Int64ToStr(cf_projects_currency_id), false)
		if err != nil {
			return p.ErrInfo(err)
		}
		// возможно были списания по кредиту
		err = p.loanPaymentsRollback(p.BlockData.UserId, cf_projects_currency_id)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	// возможно нужно откатить таблицу points_status
	err = p.pointsUpdateRollbackMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.pointsUpdateRollbackMain(p.BlockData.UserId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.mydctxRollback()
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CfSendDcRollbackFront() error {
	return p.limitRequestsRollback("cf_send_dc")
}
