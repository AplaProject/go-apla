package dcparser

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

// Эту транзакцию имеет право генерить только нод, который генерит данный блок
// подписана нодовским ключом
func (p *Parser) NewPctInit() error {
	var err error
	var fields []string
	fields = []string{"new_pct", "sign"}
	p.TxMap, err = p.GetTxMap(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

type newPctType struct {
	Currency map[string]map[string]string `json:"currency"`
	Referral map[string]int64             `json:"referral"`
}

func (p *Parser) NewPctFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	nodePublicKey, err := p.GetNodePublicKey(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(nodePublicKey) == 0 {
		return p.ErrInfo("incorrect user_id")
	}

	newPctCurrency := make(map[string]map[string]string)
	// раньше не было рефских
	if p.BlockData != nil && p.BlockData.BlockId <= 77951 {
		err = json.Unmarshal([]byte(p.TxMap["new_pct"]), &newPctCurrency)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		newPctTx := new(newPctType)
		err = json.Unmarshal([]byte(p.TxMap["new_pct"]), &newPctTx)
		if err != nil {
			return p.ErrInfo(err)
		}
		if newPctTx.Referral == nil {
			return p.ErrInfo("!Referral")
		}
		newPctCurrency = newPctTx.Currency
	}
	if len(newPctCurrency) == 0 {
		return p.ErrInfo("!newPctCurrency")
	}

	// проверим, верно ли указаны ID валют
	currencyIdsSql := ""
	countCurrency := 0
	for id := range newPctCurrency {
		currencyIdsSql += id + ","
		countCurrency++
	}
	currencyIdsSql = currencyIdsSql[0 : len(currencyIdsSql)-1]
	count, err := p.Single("SELECT count(id) FROM currency WHERE id IN (" + currencyIdsSql + ")").Int()
	if err != nil {
		return p.ErrInfo(err)
	}
	if count != countCurrency {
		return p.ErrInfo("count_currency")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["new_pct"])
	CheckSignResult, err := utils.CheckSign([][]byte{nodePublicKey}, forSign, p.TxMap["sign"], true)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// проверим, прошло ли 2 недели с момента последнего обновления pct
	pctTime, err := p.Single("SELECT max(time) FROM pct").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxTime-pctTime <= p.Variables.Int64["new_pct_period"] {
		return p.ErrInfo(fmt.Sprintf("14 days error %d - %d <= %d", p.TxTime, pctTime, p.Variables.Int64["new_pct_period"]))
	}
	// берем все голоса miner_pct
	pctVotes := make(map[int64]map[string]map[string]int64)
	rows, err := p.Query("SELECT currency_id, pct, count(user_id) as votes FROM votes_miner_pct GROUP BY currency_id, pct ORDER BY currency_id, pct ASC")
	if err != nil {
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var currency_id, votes int64
		var pct string
		err = rows.Scan(&currency_id, &pct, &votes)
		if err != nil {
			return p.ErrInfo(err)
		}
		log.Debug("newpctcurrency_id", currency_id, "pct", pct, "votes", votes)
		if len(pctVotes[currency_id]) == 0 {
			pctVotes[currency_id] = make(map[string]map[string]int64)
		}
		if len(pctVotes[currency_id]["miner_pct"]) == 0 {
			pctVotes[currency_id]["miner_pct"] = make(map[string]int64)
		}
		pctVotes[currency_id]["miner_pct"][pct] = votes
	}

	// берем все голоса user_pct
	rows, err = p.Query("SELECT currency_id, pct, count(user_id) as votes FROM votes_user_pct GROUP BY currency_id, pct ORDER BY currency_id, pct ASC")
	if err != nil {
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var currency_id, votes int64
		var pct string
		err = rows.Scan(&currency_id, &pct, &votes)
		if err != nil {
			return p.ErrInfo(err)
		}
		log.Debug("currency_id", currency_id, "pct", pct, "votes", votes)
		if len(pctVotes[currency_id]) == 0 {
			pctVotes[currency_id] = make(map[string]map[string]int64)
		}
		if len(pctVotes[currency_id]["user_pct"]) == 0 {
			pctVotes[currency_id]["user_pct"] = make(map[string]int64)
		}
		pctVotes[currency_id]["user_pct"][pct] = votes
	}

	newPct := make(map[string]map[string]map[string]string)
	newPct["currency"] = make(map[string]map[string]string)
	var userMaxKey int64
	PctArray := utils.GetPctArray()

	log.Debug("pctVotes", pctVotes)
	for currencyId, data := range pctVotes {

		currencyIdStr := utils.Int64ToStr(currencyId)
		// определяем % для майнеров
		pctArr := utils.MakePctArray(data["miner_pct"])
		log.Debug("pctArrminer_pct", pctArr, currencyId)
		key := utils.GetMaxVote(pctArr, 0, 390, 100)
		log.Debug("key", key)
		if len(newPct["currency"][currencyIdStr]) == 0 {
			newPct["currency"][currencyIdStr] = make(map[string]string)
		}
		newPct["currency"][currencyIdStr]["miner_pct"] = utils.GetPctValue(key)

		// определяем % для юзеров
		pctArr = utils.MakePctArray(data["user_pct"])
		log.Debug("pctArruser_pct", pctArr, currencyId)
		// раньше не было завимости юзерского % от майнерского
		if p.BlockData != nil && p.BlockData.BlockId <= 95263 {
			userMaxKey = 390
		} else {
			log.Debug("newPct", newPct)
			pctY := utils.ArraySearch(newPct["currency"][currencyIdStr]["miner_pct"], PctArray)
			log.Debug("newPct[currency][currencyIdStr][miner_pct]", newPct["currency"][currencyIdStr]["miner_pct"])
			log.Debug("PctArray", PctArray)
			log.Debug("miner_pct $pct_y=", pctY)
			maxUserPctY := utils.Round(utils.StrToFloat64(pctY)/2, 2)
			userMaxKey = utils.FindUserPct(int(maxUserPctY))
			log.Debug("maxUserPctY", maxUserPctY, "userMaxKey", userMaxKey, "currencyIdStr", currencyIdStr)
			// отрезаем лишнее, т.к. поиск идет ровно до макимального возможного, т.е. до miner_pct/2
			pctArr = utils.DelUserPct(pctArr, userMaxKey)
			log.Debug("pctArr", pctArr)
		}
		key = utils.GetMaxVote(pctArr, 0, userMaxKey, 100)
		log.Debug("data[user_pct]", data["user_pct"])
		log.Debug("pctArr", pctArr)
		log.Debug("userMaxKey", userMaxKey)
		log.Debug("key", key)
		newPct["currency"][currencyIdStr]["user_pct"] = utils.GetPctValue(key)
		log.Debug("user_pct", newPct["currency"][currencyIdStr]["user_pct"])
	}

	var jsonData []byte
	// раньше не было рефских
	if p.BlockData != nil && p.BlockData.BlockId <= 77951 {

		newPct_ := newPct["currency"]
		jsonData, err = json.Marshal(newPct_)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {

		newPct_ := new(newPctType)
		newPct_.Currency = make(map[string]map[string]string)
		newPct_.Currency = newPct["currency"]
		newPct_.Referral = make(map[string]int64)
		refLevels := []string{"first", "second", "third"}
		for i := 0; i < len(refLevels); i++ {
			level := refLevels[i]
			var votesReferral []map[int64]int64

			// берем все голоса
			rows, err := p.Query("SELECT " + level + ", count(user_id) as votes FROM votes_referral GROUP BY " + level + " ORDER BY " + level + " ASC ")
			if err != nil {
				return p.ErrInfo(err)
			}
			defer rows.Close()
			for rows.Next() {
				var level_, votes int64
				err = rows.Scan(&level_, &votes)
				if err != nil {
					return p.ErrInfo(err)
				}
				votesReferral = append(votesReferral, map[int64]int64{level_: votes})
			}
			newPct_.Referral[level] = (utils.GetMaxVote(votesReferral, 0, 30, 10))
		}
		jsonData, err = json.Marshal(newPct_)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	if string(p.TxMap["new_pct"]) != string(jsonData) {
		return p.ErrInfo("p.TxMap[new_pct] != jsonData " + string(p.TxMap["new_pct"]) + "!=" + string(jsonData))
	}
	log.Debug(string(jsonData))

	return nil
}

func (p *Parser) NewPct() error {

	newPctCurrency := make(map[string]map[string]string)
	newPctTx := new(newPctType)
	// раньше не было рефских
	if p.BlockData.BlockId <= 77951 {
		err := json.Unmarshal([]byte(p.TxMap["new_pct"]), &newPctCurrency)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		err := json.Unmarshal([]byte(p.TxMap["new_pct"]), &newPctTx)
		if err != nil {
			return p.ErrInfo(err)
		}
		if newPctTx.Referral == nil {
			return p.ErrInfo("!Referral")
		}
		newPctCurrency = newPctTx.Currency
	}
	for currencyId, data := range newPctCurrency {
		err := p.ExecSql("INSERT INTO pct ( time, currency_id, miner, user, block_id ) VALUES ( ?, ?, ?, ?, ? )", p.BlockData.Time, currencyId, data["miner_pct"], data["user_pct"], p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	if p.BlockData.BlockId > 77951 {
		err := p.selectiveLoggingAndUpd([]string{"first", "second", "third"}, []interface{}{utils.Int64ToStr(newPctTx.Referral["first"]), utils.Int64ToStr(newPctTx.Referral["second"]), utils.Int64ToStr(newPctTx.Referral["third"])}, "referral", []string{}, []string{})
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) NewPctRollback() error {
	if p.BlockData.BlockId > 77951 {
		err := p.selectiveRollback([]string{"first", "second", "third"}, "referral", "", false)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	affect, err := p.ExecSqlGetAffect("DELETE FROM pct WHERE block_id = ?", p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("pct", affect)
	return nil
}

func (p *Parser) NewPctRollbackFront() error {

	return nil
}
