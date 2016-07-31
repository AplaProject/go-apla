package dcparser

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"sort"
	"time"
)

func (p *Parser) VotesComplexInit() error {
	var err error
	var fields []string
	fields = []string{"json_data", "sign"}
	p.TxMap, err = p.GetTxMap(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func makeVcomplex(json_data []byte) (*vComplex, error) {
	vComplex := new(vComplex)
	err := json.Unmarshal(json_data, &vComplex)
	if err != nil {
		vComplex_ := new(vComplex_)
		err = json.Unmarshal(json_data, &vComplex_)
		if err != nil {
			vComplex__ := new(vComplex__)
			err = json.Unmarshal(json_data, &vComplex__)
			if err != nil {
				return vComplex, err
			}
			vComplex.Referral = vComplex__.Referral
			vComplex.Currency = vComplex__.Currency
			vComplex.Admin = utils.StrToInt64(vComplex__.Admin)
		} else {
			vComplex.Referral = make(map[string]string)
			for k, v := range vComplex_.Referral {
				vComplex.Referral[k] = utils.Int64ToStr(v)
			}
			vComplex.Currency = vComplex_.Currency
			vComplex.Admin = vComplex_.Admin
		}
	}
	return vComplex, nil
}

func (p *Parser) VotesComplexFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	var txTime int64
	if p.BlockData != nil {
		txTime = p.BlockData.Time
	} else {
		txTime = time.Now().Unix() - 30
	}

	// прошло ли 30 дней с момента регистрации майнера
	err = p.checkMinerNewbie()
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["json_data"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	currencyVotes := make(map[string][]float64)
	var doubleCheck []int64
	// раньше не было рефских
	if p.BlockData == nil || p.BlockData.BlockId > 77951 {

		vComplex, err := makeVcomplex(p.TxMap["json_data"])
		if err != nil {
			return p.ErrInfo(err)
		}

		if vComplex.Referral == nil {
			return p.ErrInfo("!Referral")
		}
		if vComplex.Currency == nil {
			return p.ErrInfo("!Currency")
		}
		if p.BlockData == nil || p.BlockData.BlockId > 153750 {
			if vComplex.Admin > 0 {
				adminUserId, err := p.Single("SELECT user_id FROM users WHERE user_id  =  ?", vComplex.Admin).Int64()
				if err != nil {
					return p.ErrInfo(err)
				}
				if adminUserId == 0 {
					return p.ErrInfo("incorrect admin user_id")
				}
			}
		}
		if !utils.CheckInputData(vComplex.Referral["first"], "referral") || !utils.CheckInputData(vComplex.Referral["second"], "referral") || !utils.CheckInputData(vComplex.Referral["third"], "referral") {
			return p.ErrInfo("incorrect referral")
		}
		currencyVotes = vComplex.Currency
	} else {
		vComplex := make(map[string][]float64)
		err = json.Unmarshal(p.TxMap["json_data"], &vComplex)
		if err != nil {
			return p.ErrInfo(err)
		}
		currencyVotes = vComplex
	}
	for currencyId, data := range currencyVotes {
		if !utils.CheckInputData(currencyId, "int") {
			return p.ErrInfo("incorrect currencyId")
		}

		// проверим, что нет дублей
		if utils.InSliceInt64(utils.StrToInt64(currencyId), doubleCheck) {
			return p.ErrInfo("double currencyId")
		}
		doubleCheck = append(doubleCheck, utils.StrToInt64(currencyId))

		// есть ли такая валюта
		currencyId_, err := p.Single("SELECT id FROM currency WHERE id  =  ?", currencyId).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if currencyId_ == 0 {
			return p.ErrInfo("incorrect currencyId")
		}
		// у юзера по данной валюте должна быть обещанная сумма, которая имеет статус mining/repaid и находится с таким статусом >90 дней
		id, err := p.Single("SELECT id FROM promised_amount	WHERE currency_id  =  ? AND user_id  =  ? AND status IN ('mining', 'repaid') AND start_time < ? AND start_time > 0 AND del_block_id  =  0 AND del_mining_block_id  =  0", currencyId, p.TxUserID, (txTime - p.Variables.Int64["min_hold_time_promise_amount"])).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if id == 0 {
			return p.ErrInfo("incorrect currencyId")
		}

		// если по данной валюте еще не набралось >1000 майнеров, то за неё голосовать нельзя.
		countMiners, err := p.Single(`
			SELECT count(*) FROM
			(SELECT user_id
			FROM promised_amount
			WHERE start_time < ? AND del_block_id  =  0 AND status IN ('mining', 'repaid') AND currency_id  =  ? AND del_block_id  =  0 AND del_mining_block_id  =  0
			GROUP BY user_id) as t1`, (txTime - p.Variables.Int64["min_hold_time_promise_amount"]), currencyId).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if countMiners < p.Variables.Int64["min_miners_of_voting"] {
			return p.ErrInfo("countMiners")
		}
		if len(data) != 5 {
			return p.ErrInfo("incorrect data")
		}
		if !utils.CheckPct0(data[0]) {
			return p.ErrInfo("incorrect miner_pct "+utils.Float64ToStr(data[0]))
		}
		if !utils.CheckPct0(data[1]) {
			return p.ErrInfo("incorrect user_pct "+utils.Float64ToStr(data[1]))
		}

		// max promise amount
		if !utils.InSliceInt64(int64(data[2]), utils.GetAllMaxPromisedAmount()) {
			log.Debug("%v", int64(data[2]))
			log.Debug("%v", utils.GetAllMaxPromisedAmount())
			return p.ErrInfo("incorrect max promised amount")
		}

		totalCountCurrencies, err := p.Single("SELECT count(id) FROM currency").Int64()
		if err != nil {
			return p.ErrInfo(err)
		}

		// max other currency 0/1/2/3/.../76
		if !utils.CheckInputData(int(data[3]), "int") || int64(data[3]) > totalCountCurrencies {
			return p.ErrInfo(fmt.Sprintf("incorrect max other currency %d > %d", data[3], totalCountCurrencies))
		}

		currencyCount, err := p.Single("SELECT count(id) FROM currency").Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if int64(data[3]) > (currencyCount - 1) {
			return p.ErrInfo("incorrect max other currency")
		}
		// reduction 10/25/50/90
		if !utils.InSliceInt64(int64(data[4]), consts.ReductionDC) {
			return p.ErrInfo("incorrect reduction")
		}
	}

	err = p.limitRequest(p.Variables.Int64["limit_votes_complex"], "votes_complex", p.Variables.Int64["limit_votes_complex_period"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) VotesComplex() error {

	currencyVotes := make(map[string][]float64)

	if p.BlockData.BlockId > 77951 {
		vComplex, err := makeVcomplex(p.TxMap["json_data"])
		if err != nil {
			return p.ErrInfo(err)
		}
		currencyVotes = vComplex.Currency
		// голоса за реф. %
		p.selectiveLoggingAndUpd([]string{"first", "second", "third"}, []interface{}{vComplex.Referral["first"], vComplex.Referral["second"], vComplex.Referral["third"]}, "votes_referral", []string{"user_id"}, []string{string(p.TxMap["user_id"])})

		// раньше не было выборов админа
		if p.BlockData.BlockId >= 153750 && vComplex.Admin > 0 {
			p.selectiveLoggingAndUpd([]string{"admin_user_id", "time"}, []interface{}{vComplex.Admin, p.TxTime}, "votes_admin", []string{"user_id"}, []string{string(p.TxMap["user_id"])})
		}

	} else { // раньше не было рефских и выбора админа
		vComplex := make(map[string][]float64)
		err := json.Unmarshal(p.TxMap["json_data"], &vComplex)
		if err != nil {
			return p.ErrInfo(err)
		}
		currencyVotes = vComplex
	}

	var currencyIds []int
	for k := range currencyVotes {
		currencyIds = append(currencyIds, utils.StrToInt(k))
	}
	sort.Ints(currencyIds)
	//sort.Sort(sort.Reverse(sort.IntSlice(keys)))

	for _, currencyId := range currencyIds {

		data := currencyVotes[utils.IntToStr(currencyId)]
		currencyIdStr := utils.IntToStr(currencyId)
		UserIdStr := string(p.TxMap["user_id"])
		// miner_pct
		err := p.selectiveLoggingAndUpd([]string{"pct", "time"}, []interface{}{utils.Float64ToStr(data[0]), p.TxTime}, "votes_miner_pct", []string{"user_id", "currency_id"}, []string{UserIdStr, currencyIdStr})
		if err != nil {
			return p.ErrInfo(err)
		}

		// user_pct
		err = p.selectiveLoggingAndUpd([]string{"pct"}, []interface{}{utils.Float64ToStr(data[1])}, "votes_user_pct", []string{"user_id", "currency_id"}, []string{UserIdStr, currencyIdStr})
		if err != nil {
			return p.ErrInfo(err)
		}

		// max_promised_amount
		err = p.selectiveLoggingAndUpd([]string{"amount"}, []interface{}{int64(data[2])}, "votes_max_promised_amount", []string{"user_id", "currency_id"}, []string{UserIdStr, currencyIdStr})
		if err != nil {
			return p.ErrInfo(err)
		}

		// max_other_currencies
		err = p.selectiveLoggingAndUpd([]string{"count"}, []interface{}{int64(data[3])}, "votes_max_other_currencies", []string{"user_id", "currency_id"}, []string{UserIdStr, currencyIdStr})
		if err != nil {
			return p.ErrInfo(err)
		}

		// reduction
		err = p.selectiveLoggingAndUpd([]string{"pct", "time"}, []interface{}{int64(data[4]), p.TxTime}, "votes_reduction", []string{"user_id", "currency_id"}, []string{UserIdStr, currencyIdStr})
		if err != nil {
			return p.ErrInfo(err)
		}

		// проверим, не наш ли это user_id
		myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}

		if p.TxUserID == myUserId && myBlockId <= p.BlockData.BlockId {
			// отметимся, что голосовали, чтобы не пришло уведомление о необходимости голосовать раз в 2 недели
			// может быть дубль, поэтому ошибки не проверяем
			p.ExecSql("INSERT INTO "+myPrefix+"my_complex_votes ( last_voting ) VALUES ( ? )", p.BlockData.Time)
		}
	}

	return nil
}

func (p *Parser) VotesComplexRollback() error {

	currencyVotes := make(map[string][]float64)

	if p.BlockData.BlockId > 77951 {
		vComplex, err := makeVcomplex(p.TxMap["json_data"])
		if err != nil {
			return p.ErrInfo(err)
		}
		if p.BlockData.BlockId > 153750 && vComplex.Admin != 0 {
			err := p.selectiveRollback([]string{"admin_user_id", "time"}, "votes_admin", "user_id="+string(p.TxMap["user_id"]), false)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
		// голоса за реф. %
		err = p.selectiveRollback([]string{"first", "second", "third"}, "votes_referral", "user_id="+string(p.TxMap["user_id"]), false)
		if err != nil {
			return p.ErrInfo(err)
		}

		currencyVotes = vComplex.Currency

	} else { // раньше не было рефских и выбора админа
		vComplex := make(map[string][]float64)
		err := json.Unmarshal(p.TxMap["json_data"], &vComplex)
		if err != nil {
			return p.ErrInfo(err)
		}
		currencyVotes = vComplex
	}

	// сортируем по $currency_id в обратном порядке
	var currencyIds []int
	for k := range currencyVotes {
		currencyIds = append(currencyIds, utils.StrToInt(k))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(currencyIds)))

	for _, currencyId := range currencyIds {
		currencyIdStr := utils.IntToStr(currencyId)
		// miner_pct
		err := p.selectiveRollback([]string{"pct", "time"}, "votes_miner_pct", "user_id="+string(p.TxMap["user_id"])+" AND currency_id = "+currencyIdStr, false)
		if err != nil {
			return p.ErrInfo(err)
		}
		// user_pct
		err = p.selectiveRollback([]string{"pct"}, "votes_user_pct", "user_id="+string(p.TxMap["user_id"])+" AND currency_id = "+currencyIdStr, false)
		if err != nil {
			return p.ErrInfo(err)
		}
		// reduction
		err = p.selectiveRollback([]string{"pct", "time"}, "votes_reduction", "user_id="+string(p.TxMap["user_id"])+" AND currency_id = "+currencyIdStr, false)
		if err != nil {
			return p.ErrInfo(err)
		}
		// max_promised_amount
		err = p.selectiveRollback([]string{"amount"}, "votes_max_promised_amount", "user_id="+string(p.TxMap["user_id"])+" AND currency_id = "+currencyIdStr, false)
		if err != nil {
			return p.ErrInfo(err)
		}
		// max_other_currencies
		err = p.selectiveRollback([]string{"count"}, "votes_max_other_currencies", "user_id="+string(p.TxMap["user_id"])+" AND currency_id = "+currencyIdStr, false)
		if err != nil {
			return p.ErrInfo(err)
		}

		// проверим, не наш ли это user_id
		myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}

		if p.TxUserID == myUserId {
			// отметимся, что голосовали, чтобы не пришло уведомление о необходимости голосовать раз в 2 недели
			err = p.ExecSql("DELETE FROM "+myPrefix+"my_complex_votes WHERE last_voting =?", p.BlockData.Time)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}

	return nil
}

func (p *Parser) VotesComplexRollbackFront() error {
	return p.limitRequestsRollback("votes_complex")
}
