package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"strings"
)

// если из-за смены местоположения или изначально после new_promised_amount получили rejected,
// то просто шлем новый запрос. возможно был косяк с видео-файлом.
// Если было delete=1, то перезаписываем

func (p *Parser) NewPromisedAmountInit() error {
	var fields []map[string]string
	if p.BlockData != nil && p.BlockData.BlockId < 27134 {
		fields = []map[string]string{{"currency_id": "int64"}, {"amount": "money"}, {"video_type": "string"}, {"video_url_id": "string"}, {"sign": "bytes"}}
	} else {
		fields = []map[string]string{{"currency_id": "int64"}, {"amount": "money"}, {"video_type": "string"}, {"video_url_id": "string"}, {"payment_systems_ids": "string"}, {"sign": "bytes"}}
	}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewPromisedAmountFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"currency_id": "int", "amount": "amount", "video_type": "video_type", "video_url_id": "video_url_id"}
	if p.BlockData == nil || p.BlockData.BlockId > 27134 {
		verifyData["payment_systems_ids"] = "payment_systems_ids"
	}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, существует ли такая валюта
	if ok, err := p.CheckCurrency(p.TxMaps.Int64["currency_id"]); !ok {
		return p.ErrInfo(err)
	}

	// юзер должен быть или miner, или passive_miner, т.е. иметь miner_id. не даем майнерам, которых забанил админ, добавлять новые обещанные суммы.
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим статус. должно  вообще не быть записей. всё, что rejected/change_geo и пр. юзер должен вначале удалить
	data, err := p.OneRow("SELECT status, currency_id FROM promised_amount WHERE currency_id  =  ? AND del_block_id  =  0 AND del_mining_block_id  =  0 AND user_id  =  ?", p.TxMaps.Int64["currency_id"], p.TxMaps.Int64["user_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(data["status"]) > 0 {
		return p.ErrInfo("exists promised_amount")
	}

	newMaxPromisedAmount, err := p.Single("SELECT amount FROM max_promised_amounts WHERE currency_id  =  ? ORDER BY time DESC", p.TxMaps.Int64["currency_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	newMaxOtherCurrencies, err := p.Single("SELECT max_other_currencies FROM currency WHERE id  =  ?", p.TxMaps.Int64["currency_id"]).Int()
	if err != nil {
		return p.ErrInfo(err)
	}

	// пока нет хотя бы 1000 майнеров по этой валюте, нельзя добавлять обещанную сумму не из зеленой зоны
	countMiners, err := p.Single("SELECT count(id) FROM promised_amount where currency_id = ? AND status='mining'", p.TxMaps.Int64["currency_id"]).Int64()

	if countMiners < 1000 && (p.BlockData == nil || p.BlockData.BlockId > 283749) {
		newMaxPromisedAmount = consts.MaxGreen[p.TxMaps.Int64["currency_id"]]
	}

	// т.к. можно перевести из mining в repaid, где нет лимитов, и так проделать много раз, то
	// нужно жестко лимитировать ОБЩУЮ сумму по всем promised_amount данной валюты
	repaidAmount, err := p.GetRepaidAmount(p.TxMaps.Int64["currency_id"], p.TxUserID)
	if p.TxMaps.Money["amount"]+repaidAmount > float64(newMaxPromisedAmount) {
		return p.ErrInfo("amount")
	}


	// возьмем id всех добавленных валют
	existsCurrencies, err := p.GetList("SELECT currency_id FROM promised_amount WHERE user_id  =  ? AND del_block_id  =  0 AND del_mining_block_id  =  0 GROUP BY currency_id", p.TxMaps.Int64["user_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// нельзя добавлять новую валюту, пока не одобрена хотя бы одна, т.е. пока нет WOC
	woc, err := p.Single("SELECT id FROM promised_amount WHERE user_id  =  ? AND currency_id  =  1", p.TxMaps.Int64["user_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(existsCurrencies) > 0 && woc == 0 {
		return p.ErrInfo("!$woc")
	}
	if len(existsCurrencies) > 0 {

		// можно ли новую валюту иметь с таким кол-вом валют как у нас
		if len(existsCurrencies) > newMaxOtherCurrencies {
			return p.ErrInfo(fmt.Sprintf("max_other_currencies (%d > %d)", len(existsCurrencies), newMaxOtherCurrencies))
		}

		// проверим, можно ли к существующим валютам добавить новую
		for _, currencyId := range existsCurrencies {
			maxOtherCurrencies, err := p.Single("SELECT max_other_currencies FROM currency WHERE id  =  ?", currencyId).Int()
			if err != nil {
				return p.ErrInfo(err)
			}
			if len(existsCurrencies) > maxOtherCurrencies {
				return p.ErrInfo("max_other_currencies")
			}
		}
	}
	// должно быть geolocation
	latitude, err := p.Single("SELECT latitude FROM miners_data WHERE user_id  =  ?", p.TxMaps.Int64["user_id"]).Float64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if latitude == 0 && p.TxUserID != 1 {
		return p.ErrInfo("!geo")
	}
	/*
		var txTime int64
		if p.BlockData!=nil { // тр-ия пришла в блоке
			txTime = p.BlockData.Time
		} else { // голая тр-ия
			txTime = time.Now().Unix() - 30 // просто на всякий случай небольшой запас
		}*/
	err = p.CheckCashRequests(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := ""
	if p.BlockData != nil && p.BlockData.BlockId < 27134 {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["currency_id"], p.TxMap["amount"], p.TxMap["video_type"], p.TxMap["video_url_id"])
	} else {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["currency_id"], p.TxMap["amount"], p.TxMap["video_type"], p.TxMap["video_url_id"], p.TxMap["payment_systems_ids"])
	}
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(p.Variables.Int64["limit_promised_amount"], "promised_amount", p.Variables.Int64["limit_promised_amount_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewPromisedAmount() error {
	addSqlNames := ""
	addSqlValues := ""
	if p.BlockData.BlockId > 27134 {
		paymentSystemsIds := strings.Split(string(p.TxMaps.String["payment_systems_ids"]), ",")
		for i, v := range paymentSystemsIds {
			addSqlNames += fmt.Sprintf("ps%d,", (i + 1))
			addSqlValues += fmt.Sprintf("%s,", v)
		}
	}

	var profit float64
	var tdcAmountUpdate int64
	if p.BlockData.BlockId < 310000 { // перенсено в mining_sn
		// возможно юзер хочет забрать ранее намайненное когда он еще не был майнером
		restrictedPA, err := p.OneRow(`SELECT * from promised_amount_restricted WHERE currency_id = ? AND user_id = ?`, p.TxMaps.Int64["currency_id"], p.TxUserID).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		// провереим, точно ли это первый перевод из урезанных
		existsPA, err := p.Single(`SELECT id from promised_amount WHERE currency_id = ? AND user_id = ?`, p.TxMaps.Int64["currency_id"], p.TxUserID).Int64()
		if existsPA == 0 {
			pct, err := p.GetPct()
			if err != nil {
				return p.ErrInfo(err)
			}
			startTime := utils.StrToInt64(restrictedPA["last_update"])
			// максимум 6 мес. ~ 26$
			endTime := p.BlockData.Time
			if p.BlockData.Time - startTime > 86400*180 {
				endTime = startTime + 86400*180
			}
			profit, err = p.calcProfit_(utils.StrToFloat64(restrictedPA["amount"]), startTime, endTime, pct[p.TxMaps.Int64["currency_id"]], []map[int64]string{{0: "user"}}, [][]int64{}, []map[int64]string{}, 0, 0)
			if err != nil {
				return p.ErrInfo(err)
			}
			tdcAmountUpdate = p.BlockData.Time
		}
	}
	//добавляем promised_amount в БД
	err := p.ExecSql(`
				INSERT INTO promised_amount (
						user_id,
						amount,
						currency_id,
						tdc_amount,
						tdc_amount_update,
						` + addSqlNames + `
						video_type,
						video_url_id,
						votes_start_time
					)
					VALUES (
						` + utils.Int64ToStr(p.TxMaps.Int64["user_id"]) + `,
						` + utils.Float64ToStr(p.TxMaps.Money["amount"]) + `,
						` + utils.Int64ToStr(p.TxMaps.Int64["currency_id"]) + `,
						` + utils.Float64ToStr(profit) + `,
						` + utils.Int64ToStr(tdcAmountUpdate) + `,
						` + addSqlValues + `
						'` + p.TxMaps.String["video_type"] + `',
						'` + p.TxMaps.String["video_url_id"] + `',
						` + utils.Int64ToStr(p.BlockData.Time) + `
					)`)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, не наш ли это user_id
	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxMaps.Int64["user_id"])
	if err != nil {
		return err
	}
	if p.TxUserID == myUserId && myBlockId <= p.BlockData.BlockId {
		// Удалим, т.к. попало в блок
		err = p.ExecSql("DELETE FROM "+myPrefix+"my_promised_amount WHERE amount = ? AND currency_id = ?", p.TxMaps.Money["amount"], p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) NewPromisedAmountRollback() error {
	err := p.ExecSql("DELETE FROM promised_amount WHERE user_id = ? AND amount = ? AND currency_id = ? AND status = 'pending' AND votes_start_time = ?", p.TxMaps.Int64["user_id"], p.TxMaps.Money["amount"], p.TxMaps.Int64["currency_id"], p.BlockData.Time)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("promised_amount", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewPromisedAmountRollbackFront() error {
	return p.limitRequestsRollback("promised_amount")
}
