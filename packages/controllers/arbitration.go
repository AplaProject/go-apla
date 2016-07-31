package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type arbitrationPage struct {
	SignData              string
	ShowSignData          bool
	TxType                string
	TxTypeId              int64
	TimeNow               int64
	UserId                int64
	Alert                 string
	Lang                  map[string]string
	CountSignArr          []int
	Arbitrators           []*arbitrationType
	MyTrustList           []map[string]string
	PendingTx             int64
	Arbitrator            int64
	ArbitrationDaysRefund int64
	LastTxFormatted       string
	ArbitrationTrustList  int64
}

type arbitrationType struct {
	Arbitrator_user_id     int64
	Url                    string
	Count                  int64
	Count_rejected_refunds int64
	Refund_data_count      int64
	Refund_data_sum        float64
}

func (c *Controller) Arbitration() (string, error) {

	txType := "ChangeArbitratorList"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	var q string
	if c.ConfigIni["db_type"] == "postgresql" {
		q = `SELECT DISTINCT arbitrator_user_id, url, count(arbitration_trust_list.user_id) as count
			FROM arbitration_trust_list
			LEFT JOIN users ON users.user_id = arbitration_trust_list.arbitrator_user_id
			WHERE arbitration_trust_list.user_id = ? AND
						 arbitration_trust_list.arbitrator_user_id > 0
			GROUP BY arbitrator_user_id, url
			ORDER BY count(arbitration_trust_list.user_id)  DESC`
	} else {
		q = `SELECT arbitrator_user_id, url, count(arbitration_trust_list.user_id) as count
			FROM arbitration_trust_list
			LEFT JOIN users ON users.user_id = arbitration_trust_list.arbitrator_user_id
			WHERE arbitration_trust_list.user_id = ? AND
						 arbitration_trust_list.arbitrator_user_id > 0
			GROUP BY arbitrator_user_id
			ORDER BY count(arbitration_trust_list.user_id)  DESC`
	}
	myTrustList, err := c.GetAll(q, 100, c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	var arbitrators []*arbitrationType
	// top 10 арбитров
	if c.ConfigIni["db_type"] == "postgresql" {
		q = `SELECT DISTINCT arbitrator_user_id, url, count(arbitration_trust_list.user_id) as count
			FROM arbitration_trust_list
			LEFT JOIN miners_data ON miners_data.user_id = arbitration_trust_list.user_id
			LEFT JOIN users ON users.user_id = arbitration_trust_list.arbitrator_user_id
			WHERE miners_data.status='miner' AND
						 arbitration_trust_list.arbitrator_user_id > 0
			GROUP BY arbitrator_user_id, url
			ORDER BY count(arbitration_trust_list.user_id)  DESC
			LIMIT 10`
	} else {
		q = `SELECT arbitrator_user_id, url, count(arbitration_trust_list.user_id) as count
			FROM arbitration_trust_list
			LEFT JOIN miners_data ON miners_data.user_id = arbitration_trust_list.user_id
			LEFT JOIN users ON users.user_id = arbitration_trust_list.arbitrator_user_id
			WHERE miners_data.status='miner' AND
						 arbitration_trust_list.arbitrator_user_id > 0
			GROUP BY arbitrator_user_id
			ORDER BY count(arbitration_trust_list.user_id)  DESC
			LIMIT 10`
	}
	rows, err := c.Query(c.FormatQuery(q))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		//var pct, amount float64
		var arbitrator_user_id, count int64
		var url string
		err = rows.Scan(&arbitrator_user_id, &url, &count)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		// кол-во манибеков и сумма за последний месяц
		refund_data, err := c.OneRow(`
			SELECT count(id) as count, sum(refund) as sum
			FROM orders
			LEFT JOIN miners_data ON miners_data.user_id = orders.buyer
			WHERE (arbitrator0 = ? OR arbitrator1 = ? OR arbitrator2 = ? OR arbitrator3 = ? OR arbitrator4 = ?) AND
						 orders.status = 'refund' AND
						 arbitrator_refund_time > ? AND
						 arbitrator_refund_time < ? AND
						 miners_data.status = 'miner'
			GROUP BY user_id
		`, arbitrator_user_id, arbitrator_user_id, arbitrator_user_id, arbitrator_user_id, arbitrator_user_id, timeNow-3600*24*30, timeNow).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		// кол-во неудовлетвореных манибеков за последний месяц
		count_rejected_refunds, err := c.Single(`
			SELECT count(id)
			FROM orders
			LEFT JOIN miners_data ON miners_data.user_id = orders.buyer
			WHERE  (arbitrator0 = ? OR arbitrator1 = ? OR arbitrator2 = ? OR arbitrator3 = ? OR arbitrator4 = ?) AND
						 orders.status = 'refund' AND
						 end_time > ? AND
						 end_time < ? AND
						 voluntary_refund = 0 AND
						 refund = 0 AND
						 miners_data.status = 'miner'
			GROUP BY user_id
		`, arbitrator_user_id, arbitrator_user_id, arbitrator_user_id, arbitrator_user_id, arbitrator_user_id, timeNow-3600*24*30, timeNow).Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		log.Debug("utils.StrToInt64(refund_data[count])", utils.StrToInt64(refund_data["count"]))
		log.Debug("utils.StrToInt64(refund_data[sum])", utils.StrToInt64(refund_data["sum"]))

		arbitrators = append(arbitrators, &arbitrationType{Arbitrator_user_id: arbitrator_user_id, Url: url, Count: count, Refund_data_count: utils.StrToInt64(refund_data["count"]), Refund_data_sum: utils.StrToFloat64(refund_data["sum"]), Count_rejected_refunds: count_rejected_refunds})

	}

	// арбитр ли наш юзер
	arbitrator, err := c.Single("SELECT conditions FROM arbitrator_conditions WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// продавец ли
	arbitrationDaysRefund, err := c.Single("SELECT arbitration_days_refund FROM users WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	arbitrationTrustList, err := c.Single("SELECT arbitrator_user_id FROM arbitration_trust_list WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangeArbitratorList", "ChangeArbitratorConditions", "ChangeSellerHoldBack", "ChangeSellerHoldBack", "MoneyBackRequest", "MoneyBack", "ChangeMoneyBackTime"}), 3, c.TimeFormat)
	lastTxFormatted := ""
	var pendingTx_ map[int64]int64
	if len(last_tx) > 0 {
		lastTxFormatted, pendingTx_ = utils.MakeLastTx(last_tx, c.Lang)
	}
	pendingTx := pendingTx_[txTypeId]

	TemplateStr, err := makeTemplate("arbitration", "arbitration", &arbitrationPage{
		Alert:                 c.Alert,
		Lang:                  c.Lang,
		CountSignArr:          c.CountSignArr,
		ShowSignData:          c.ShowSignData,
		UserId:                c.SessUserId,
		TimeNow:               timeNow,
		TxType:                txType,
		TxTypeId:              txTypeId,
		SignData:              "",
		Arbitrators:           arbitrators,
		MyTrustList:           myTrustList,
		PendingTx:             pendingTx,
		Arbitrator:            arbitrator,
		ArbitrationDaysRefund: arbitrationDaysRefund,
		LastTxFormatted:       lastTxFormatted,
		ArbitrationTrustList:  arbitrationTrustList})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
