package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
	//"github.com/blevesearch/bleve/search/searchers"
)

type AssignmentsPage struct {
	Alert              string
	SignData           string
	ShowSignData       bool
	TxType             string
	TxTypeId           int64
	TimeNow            int64
	UserId             int64
	Lang               map[string]string
	CountSignArr       []int
	CurrencyList       map[int64]string
	CashRequestsStatus map[string]string
	MyCashRequests     []map[string]string
	ActualData         map[string]string
	MainQuestion       string
	NewPromiseAmount   string
	MyRace             string
	MyCountry          string
	ExamplePoints      map[string]string
	VideoHost          string
	PhotoHosts         []string
	PromisedAmountData map[string]string
	UserInfo           map[string]string
	CloneHosts         map[int64][]string
	SN string
	SnUserId int64
	Voted              map[string]int64
}

func getMyCountryRace( c *Controller ) ( country int, race int64 ) {
	if data, err := c.OneRow("SELECT race, country FROM " + c.MyPrefix + "my_table").Int64(); err == nil {
		if data["race"] > 0 {
				race = data["race"]
		}
		if data["country"] > 0 {
			country = int(data["country"])
		}
	}
	return 	
}

func (c *Controller) Assignments() (string, error) {

	var randArr []int64
	// Нельзя завершить голосование юзеров раньше чем через сутки, даже если набрано нужное кол-во голосов.
	// В голосовании нодов ждать сутки не требуется, т.к. там нельзя поставить поддельных нодов


	// Модерация новых майнеров
	// берем тех, кто прошел проверку нодов (type='node_voting')
	
	getCount := func( query, qtype string ) (ret int64, err error)  {
		vid := `v.id`
		if qtype == `sn` {
			vid = `v.user_id`
		}
		if c.SessRestricted == 0 {
			ret, err = c.Single( query +
				` AND `+vid+` NOT IN ( SELECT id FROM `+c.MyPrefix+`my_tasks WHERE type=? AND time > ?)`, qtype,  utils.Time()-consts.ASSIGN_TIME ).Int64()
		} else {
			ret, err = c.Single( query ).Int64()
		}				
		if err != nil {
			return 
		}
		return
	}
	
	country, race := getMyCountryRace(c)
	
	query := ` `
	where := `WHERE votes_end  =  0 AND v.type  =  'user_voting' `
	if race > 0 || country > 0 {
		query += `left join faces as f on f.user_id=v.user_id `
		if race > 0 {
			where += fmt.Sprintf( `AND f.race=%d `, race )
		}
		if country > 0 {
			where += fmt.Sprintf( `AND f.country=%d `, country )
		}
	}

	num, err := getCount( `SELECT count(v.id) FROM votes_miners as v ` + query + where, `miner` )
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if num > 0 {
		randArr = append(randArr, 1)
	}

	// Модерация promised_amount
	// вначале получим ID валют, которые мы можем проверять.
	currency, err := c.GetList("SELECT currency_id FROM promised_amount WHERE status IN ('mining', 'repaid') AND user_id = ?", c.SessUserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	addSqlCurrency := ""
	currencyIds := strings.Join(currency, ",")
	if len(currencyIds) > 0 || c.SessUserId == 1 {
		if c.SessUserId != 1 {
			addSqlCurrency = "AND currency_id IN (" + currencyIds + ")"
		}
		num, err := getCount("SELECT count(id) FROM promised_amount as v WHERE status  =  'pending' AND del_block_id  =  0 " + addSqlCurrency, `promised_amount` )
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if num > 0 {
			randArr = append(randArr, 2)
		}
	}
	// модерация акков в соц. сетях
	/*mySnType, err := c.Single("SELECT sn_type FROM users WHERE user_id = ?", c.SessUserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}*/
	addSql := ` AND v.sn_url_id != '' AND v.user_id != ` + utils.Int64ToStr( c.SessUserId )
	/*if c.SessUserId!=1 && len(mySnType)>0 {
		addSql = ` AND sn_type = "`+mySnType+`"`
	}*/
	num,err = getCount( `SELECT count(v.user_id) FROM users as v WHERE v.status  =  'user'` + addSql, `sn` );
	if err != nil {
		return "", utils.ErrInfo(err)
	}

//	num, err = c.Single(`SELECT count(user_id) FROM users WHERE status  =  'user'` + addSql +
//  ` AND user_id NOT IN ( SELECT id FROM `+c.MyPrefix+`my_tasks WHERE type=? AND time > ?)`, `sn`,  utils.Time()-consts.ASSIGN_TIME ).Int64()
	if num > 0 {
		randArr = append(randArr, 3)
	}

	log.Debug("randArr %v", randArr)

	var AssignType int64
	if len(randArr) > 0 {
		AssignType = randArr[utils.RandInt(0, len(randArr))]
	}

	cloneHosts := make(map[int64][]string)
	var photoHosts []string
	examplePoints := make(map[string]string)
	tplName := "assignments"
	tplTitle := "assignments"
	
	voted := make(map[string]int64)
	var txType string
	var txTypeId int64
	var timeNow int64
	var snUserId int64
	var myRace, myCountry, mainQuestion, newPromiseAmount, videoHost string
	var promisedAmountData, userInfo, usersSN map[string]string
	var sn, mainQuery string
	switch AssignType {
	case 1:

		// ***********************************
		// задания по модерации новых майнеров
		// ***********************************
		txType = "VotesMiner"
		txTypeId = utils.TypeInt(txType)
		timeNow = utils.Time()
		mainQuery = `SELECT miners_data.user_id,
						 v.id as vote_id,
						 face_coords,
						 profile_coords,
						 video_type,
						 video_url_id,
						 photo_block_id,
						 photo_max_miner_id,
						 miners_keepers,
						 http_host,
						 pool_user_id
			FROM votes_miners as v ` + query +
			`LEFT JOIN miners_data ON miners_data.user_id = v.user_id ` + where
		if c.SessRestricted == 0 {
			userInfo, err = c.OneRow( mainQuery+ 
			` AND v.id NOT IN ( SELECT id FROM `+c.MyPrefix+`my_tasks WHERE type='miner' AND time > ?)`,  utils.Time()-consts.ASSIGN_TIME ).String()
		} else {
			userInfo, err = c.OneRow( mainQuery ).String()
		}
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(userInfo) == 0 {
			tplName = "assignments"
			break
		}

		if userInfo["pool_user_id"] != "0" {
			userInfo["http_host"], err = c.Single(`SELECT http_host FROM miners_data WHERE user_id = ?`, userInfo["pool_user_id"]).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}

		examplePoints, err = c.GetPoints(c.Lang)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		// получим ID майнеров, у которых лежат фото нужного нам юзера
		minersIds := utils.GetMinersKeepers(userInfo["photo_block_id"], userInfo["photo_max_miner_id"], userInfo["miners_keepers"], true)
		if len(minersIds) > 0 {
			photoHosts, err = c.GetList("SELECT CASE WHEN m.pool_user_id > 0 then (SELECT http_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE http_host END as http_host FROM miners_data as m WHERE m.miner_id IN (" + utils.JoinInts(minersIds, ",") + ")").String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}

		// отрезки майнера, которого проверяем
		relations, err := c.OneRow("SELECT * FROM faces WHERE user_id  =  ?", userInfo["user_id"]).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		// получим допустимые расхождения между точками и совместимость версий
		data_, err := c.OneRow("SELECT tolerances, compatibility FROM spots_compatibility").String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		tolerances := make(map[string]map[string]string)
		if err := json.Unmarshal([]byte(data_["tolerances"]), &tolerances); err != nil {
			return "", utils.ErrInfo(err)
		}
		var compatibility []int
		if err := json.Unmarshal([]byte(data_["compatibility"]), &compatibility); err != nil {
			return "", utils.ErrInfo(err)
		}

		// формируем кусок SQL-запроса для соотношений отрезков
		addSqlTolerances := ""
		typesArr := []string{"face", "profile"}
		for i := 0; i < len(typesArr); i++ {
			for j := 1; j <= len(tolerances[typesArr[i]]); j++ {
				currentRelations := utils.StrToFloat64(relations[typesArr[i][:1]+utils.IntToStr(j)])
				diff := utils.StrToFloat64(tolerances[typesArr[i]][utils.IntToStr(j)]) * currentRelations
				if diff == 0 {
					continue
				}
				min := currentRelations - diff
				max := currentRelations + diff
				addSqlTolerances += typesArr[i][:1] + utils.IntToStr(j) + ">" + utils.Float64ToStr(min) + " AND " + typesArr[i][:1] + utils.IntToStr(j) + " < " + utils.Float64ToStr(max) + " AND "
			}
		}
		addSqlTolerances = addSqlTolerances[:len(addSqlTolerances)-4]

		// формируем кусок SQL-запроса для совместимости версий
		addSqlCompatibility := ""
		for i := 0; i < len(compatibility); i++ {
			addSqlCompatibility += fmt.Sprintf(`%d,`, compatibility[i])
		}
		addSqlCompatibility = addSqlCompatibility[:len(addSqlCompatibility)-1]

		// получаем из БД похожие фото
		rows, err := c.Query(c.FormatQuery(`
				SELECT miners_data.user_id,
							 photo_block_id,
							 photo_max_miner_id,
							 miners_keepers
				FROM faces
				LEFT JOIN miners_data ON
						miners_data.user_id = faces.user_id
				WHERE `+addSqlTolerances+` AND
							version IN (`+addSqlCompatibility+`) AND
				             faces.status = 'used' AND
				             miners_data.user_id != ?
				LIMIT 0,100
				`), userInfo["user_id"])
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		defer rows.Close()
		for rows.Next() {
			var photo_block_id, photo_max_miner_id, miners_keepers string
			var user_id int64
			err = rows.Scan(&user_id, &photo_block_id, &photo_max_miner_id, &miners_keepers)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			// майнеры, у которых можно получить фото нужного нам юзера
			minersIds := utils.GetMinersKeepers(photo_block_id, photo_max_miner_id, miners_keepers, true)
			if len(minersIds) > 0 {
				photoHosts, err = c.GetList("SELECT CASE WHEN m.pool_user_id > 0 then (SELECT http_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE http_host END as http_host FROM miners_data as m WHERE m.miner_id  IN (" + utils.JoinInts(minersIds, ",") + ")").String()
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			}
			cloneHosts[user_id] = photoHosts
		}

		if race > 0 {
			myRace = c.Races[race]
		}
		if country > 0 {
			myCountry = consts.Countries[ country ]
		}
		voted,err = c.OneRow(`select votes_0, votes_1 from votes_miners where id=?`, userInfo["vote_id"]).Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		tplName = "assignments_new_miner"
		tplTitle = "assignmentsNewMiner"

	case 2:
		mainQuery = `SELECT id, currency_id,
							 amount,
							 user_id,
							 video_type,
							 video_url_id
				FROM promised_amount
				WHERE status =  'pending' AND
							 del_block_id = 0
				` + addSqlCurrency
		if c.SessRestricted == 0 {
			promisedAmountData, err = c.OneRow( mainQuery + 
			 ` AND id NOT IN ( SELECT id FROM `+c.MyPrefix+`my_tasks WHERE type='promised_amount' AND time > ?)`,  utils.Time()-consts.ASSIGN_TIME ).String()
		} else {
			promisedAmountData, err = c.OneRow( mainQuery ).String()
		}
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		promisedAmountData["currency_name"] = c.CurrencyList[utils.StrToInt64(promisedAmountData["currency_id"])]

		log.Debug("promisedAmountData %v", promisedAmountData)
		// если нету видео на ютубе, то получаем host юзера, где брать видео
		if promisedAmountData["video_url_id"] == "null" {
			videoHost, err = c.Single("SELECT CASE WHEN m.pool_user_id > 0 then (SELECT http_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE http_host end FROM miners_data as m WHERE user_id  =  ?", promisedAmountData["user_id"]).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}

		// каждый раз обязательно проверяем, где находится юзер
		userInfo, err = c.OneRow(`
				SELECT latitude,
							 user_id,
							 longitude,
							 photo_block_id,
							 photo_max_miner_id,
							 miners_keepers,
							 http_host,
							 pool_user_id
				FROM miners_data
				WHERE user_id = ?
				`, promisedAmountData["user_id"]).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		if userInfo["pool_user_id"] != "0" {
			userInfo["http_host"], err = c.Single(`SELECT http_host FROM miners_data WHERE user_id = ?`, userInfo["pool_user_id"]).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}

		// получим ID майнеров, у которых лежат фото нужного нам юзера
		minersIds := utils.GetMinersKeepers(userInfo["photo_block_id"], userInfo["photo_max_miner_id"], userInfo["miners_keepers"], true)
		if len(minersIds) > 0 {
			photoHosts, err = c.GetList("SELECT CASE WHEN m.pool_user_id > 0 then (SELECT http_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE http_host end FROM miners_data as m WHERE m.miner_id   IN (" + utils.JoinInts(minersIds, ",") + ")").String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
		voted,err = c.OneRow(`select votes_0, votes_1 from promised_amount where id=?`, promisedAmountData["id"]).Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		txType = "VotesPromisedAmount"
		txTypeId = utils.TypeInt(txType)
		timeNow = utils.Time()

		newPromiseAmount = strings.Replace(c.Lang["new_promise_amount"], "[amount]", promisedAmountData["amount"], -1)
		newPromiseAmount = strings.Replace(newPromiseAmount, "[currency]", promisedAmountData["currency_name"], -1)

		mainQuestion = strings.Replace(c.Lang["main_question"], "[amount]", promisedAmountData["amount"], -1)
		mainQuestion = strings.Replace(mainQuestion, "[currency]", promisedAmountData["currency_name"], -1)

		tplName = "assignments_promised_amount"
		tplTitle = "assignmentsPromisedAmount"
	case 3:

		/*mySnType, err := c.Single("SELECT sn_type FROM users WHERE user_id = ?", c.SessUserId).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}*/
		addSql = ` AND sn_url_id != ""`
		/*if c.SessUserId!=1 && len(mySnType)>0 {
			addSql = ` AND sn_type = "`+mySnType+`"`
		}*/
		orderBy := " ORDER BY RAND() "
		if c.ConfigIni["db_type"] == "sqlite" {
			orderBy = "ORDER BY RANDOM()"
		}
		mainQuery = `SELECT user_id, sn_type, sn_url_id FROM users WHERE status  =  'user'` + addSql
		if c.SessRestricted == 0 {
			usersSN, err = c.OneRow( mainQuery +
							` AND user_id NOT IN ( SELECT id FROM `+c.MyPrefix+`my_tasks WHERE type = ? AND time > ?) `+orderBy+` LIMIT 1`, `sn`,  utils.Time()-consts.ASSIGN_TIME ).String()
		} else {
			usersSN, err = c.OneRow( mainQuery ).String()
		}							
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		types := map[string]string{"fb" : "facebook.com", "vk" : "vk.com"}
		if len(usersSN)>0 {
			url := `http://`+types[usersSN["sn_type"]]+`/`+usersSN["sn_url_id"]
			sn = `<a href="` + url + `" onclick='THRUST.remote.send("` + url + `")' target="blank">`+ url +`</a>`
		}
		snUserId = utils.StrToInt64(usersSN["user_id"])
		voted,err = c.OneRow(`select votes_0, votes_1 from users where user_id=?`, snUserId).Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		
		txType = "VotesSnUser"
		txTypeId = utils.TypeInt(txType)
		timeNow = utils.Time()
		
		tplName = "assignments_sn"
		tplTitle = "assignmentsSn"

	default:
		tplName = "assignments"
		tplTitle = "assignments"
	}

	TemplateStr, err := makeTemplate(tplName, tplTitle, &AssignmentsPage{
		Alert:              c.Alert,
		Lang:               c.Lang,
		CountSignArr:       c.CountSignArr,
		ShowSignData:       c.ShowSignData,
		UserId:             c.SessUserId,
		TimeNow:            timeNow,
		TxType:             txType,
		TxTypeId:           txTypeId,
		SignData:           "",
		CurrencyList:       c.CurrencyList,
		MainQuestion:       mainQuestion,
		NewPromiseAmount:   newPromiseAmount,
		MyRace:             myRace,
		MyCountry:          myCountry,
		ExamplePoints:      examplePoints,
		VideoHost:          videoHost,
		PhotoHosts:         photoHosts,
		PromisedAmountData: promisedAmountData,
		UserInfo:           userInfo,
		Voted:              voted,
		SN:           sn,
		SnUserId: snUserId,
		CloneHosts:         cloneHosts})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
