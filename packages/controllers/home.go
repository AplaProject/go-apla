package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/stat"
	"math"
	"strings"
	"time"
)

type ChartCur struct {
	Currency  string
	Promised  string
	Dc        string
}

type homePage struct {
	Community             bool
	Lang                  map[string]string
	Title                 string
	Msg                   string
	Alert                 string
	MyNotice              map[string]string
	PoolAdmin             bool
	UserId                int64
	CashRequests          int64
	LastCashRequests	  []map[string]string
	RefPhotos 			  map[int64][]string
	ShowMap               bool
	BlockId               int64
	ConfirmedBlockId      int64
	CurrencyList          map[int64]string
	Assignments           int64
	SumPromisedAmount     map[string]string
	RandMiners            []int64
	Points                int64
	SessRestricted        int64
	PromisedAmountListGen map[int]utils.DCAmounts
	Wallets               map[int]utils.DCAmounts
	SumWallets            map[int]float64
	CurrencyPct           map[int]CurrencyPct
	Admin                 bool
	CalcTotal             float64
	CountSign             int
	CountSignArr          []int
	SignData              string
	ShowSignData          bool
	IOS                   bool
	Token                 string
	Mobile                bool
	MyChatName            string
	ExchangeUrl           string
	Miner                 bool
	ChatEnabled           string
	TopExMap              map[int64]*topEx
	ListBalance           *stat.ListBalance
	StatDays              int
	MyDcTransactions     []map[string]string	
	Names                map[string]string	
	Chart					string
	Charts               []ChartCur
	DCTarget int64
}

func (c *Controller) Home() (string, error) {

	log.Debug("first_select: %v", c.Parameters["first_select"])
	if c.Parameters["first_select"] == "1" {
		c.ExecSql(`UPDATE ` + c.MyPrefix + `my_table SET first_select=1`)
	}

	var publicKey []byte
	var poolAdmin bool
	var cashRequests int64
	var showMap bool
	if c.SessRestricted == 0 {
		var err error
		publicKey, err = c.GetMyPublicKey(c.MyPrefix)
		if err != nil {
			return "", err
		}
		publicKey = utils.BinToHex(publicKey)
		cashRequests, err = c.Single("SELECT count(id) FROM cash_requests WHERE to_user_id  =  ? AND status  =  'pending' AND for_repaid_del_block_id  =  0 AND del_block_id  =  0 and time > ?", c.SessUserId, utils.Time()-c.Variables.Int64["cash_request_time"]).Int64()
		fmt.Println("cashRequests", cashRequests)
		if err != nil {
			return "", err
		}
		show, err := c.Single("SELECT show_map FROM " + c.MyPrefix + "my_table").Int64()
		if err != nil {
			return "", err
		}
		if show > 0 {
			showMap = true
		}
	}
	if c.Community {
		poolAdminUserId, err := c.GetPoolAdminUserId()
		if err != nil {
			return "", err
		}
		if c.SessUserId == poolAdminUserId {
			poolAdmin = true
		}
	}

	wallets, err := c.GetBalances(c.SessUserId)
	if err != nil {
		return "", err
	}
	//var walletsByCurrency map[string]map[string]string
	walletsByCurrency := make(map[int]utils.DCAmounts)
	for _, data := range wallets {
		walletsByCurrency[int(data.CurrencyId)] = data
	}
	blockId, err := c.GetBlockId()
	if err != nil {
		return "", err
	}
	confirmedBlockId, err := c.GetConfirmedBlockId()
	if err != nil {
		return "", err
	}
	currencyList, err := c.GetCurrencyList(true)
	if err != nil {
		return "", err
	}
	for k, v := range currencyList {
		currencyList[k] = "d" + v
	}
	currencyList[1001] = "USD"

	// задания
	var assignments int64

	getCount := func( query, qtype string ) (err error)  {
		var ret int64
		
		vid := `v.id`
		if qtype == `sn` {
			vid = `v.user_id`
		}
		if c.SessRestricted == 0 {
			ret, err = c.Single( query +
				` AND `+vid+` NOT IN ( SELECT id FROM `+c.MyPrefix+`my_tasks WHERE type=? AND time > ?)`, qtype, time.Now().Unix()-consts.ASSIGN_TIME ).Int64()
		} else {
			ret, err = c.Single( query ).Int64()
		}				
		if err != nil {
			return 
		}
		assignments += ret
		return
	}

	query := `SELECT count(v.id) FROM votes_miners as v `
	where := `WHERE votes_end  =  0 AND v.type  =  'user_voting' `
	country, race := getMyCountryRace(c)
	if race > 0 || country > 0 {
		query += `left join faces as f on f.user_id=v.user_id `
		if race > 0 {
			where += fmt.Sprintf( `AND f.race=%d `, race )
		}
		if country > 0 {
			where += fmt.Sprintf( `AND f.country=%d `, country )
		}
	}
	if err := getCount( query + where, `miner` ); err !=nil {
		return "", err
	}

	// вначале получим ID валют, которые мы можем проверять.
	currencyIds, err := c.GetList("SELECT currency_id FROM promised_amount WHERE status IN ('mining', 'repaid') AND user_id  =  ?", c.SessUserId).String()
	if len(currencyIds) > 0 || c.SessUserId == 1 {
		addSql := ""
		if c.SessUserId == 1 {
			addSql = ""
		} else {
			addSql = "AND currency_id IN (" + strings.Join(currencyIds, ",") + ")"
		}
		if err := getCount(`SELECT count(id) FROM promised_amount as v WHERE status  =  'pending' AND del_block_id  =  0 ` + addSql,
	    	                `promised_amount` ); err !=nil {
			return "", err
		}
	}
	// модерация акков в соц. сетях
/*	mySnType, err := c.Single("SELECT sn_type FROM users WHERE user_id = ?", c.SessUserId).String()
	if err != nil {
		return "", err
	}*/
	addSql := ` AND v.sn_url_id != '' AND v.user_id != ` + utils.Int64ToStr( c.SessUserId )
/*	if c.SessUserId!=1 && len(mySnType)>0 {
		addSql = ` AND sn_type = "`+mySnType+`"`
	}*/
	
	if err := getCount( `SELECT count(v.user_id) FROM users as v WHERE v.status  =  'user'` + addSql, `sn` ); err !=nil {
		return "", err
	}

	// баллы
	points, err := c.Single("SELECT points FROM points WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", err
	}

	currency_pct := make(map[int]CurrencyPct)
	// проценты
	listPct, err := c.GetMap("SELECT * FROM currency", "id", "name")
	for id, name := range listPct {
		pct, err := c.OneRow("SELECT * FROM pct WHERE currency_id  =  ? ORDER BY block_id DESC", id).Float64()
		if err != nil {
			return "", err
		}
		currency_pct[utils.StrToInt(id)] = CurrencyPct{Name: name, Miner: (utils.Round((math.Pow(1+pct["miner"], 3600*24*365)-1)*100, 2)), User: (utils.Round((math.Pow(1+pct["user"], 3600*24*365)-1)*100, 2)), MinerBlock: (utils.Round((math.Pow(1+pct["miner"], 120)-1)*100, 4)), UserBlock: (utils.Round((math.Pow(1+pct["user"], 120)-1)*100, 4)), MinerSec: (pct["miner"]), UserSec: (pct["user"])}
	}
	// случайне майнеры для нанесения на карту
	maxMinerId, err := c.Single("SELECT max(miner_id) FROM miners_data").Int64()
	if err != nil {
		return "", err
	}
	randMiners, err := c.GetList("SELECT user_id FROM miners_data WHERE status  =  'miner' AND user_id > 7 AND user_id != 106 AND longitude > 0 AND miner_id IN (" + strings.Join(utils.RandSlice(1, maxMinerId, 3), ",") + ") LIMIT 3").Int64()
	if err != nil {
		return "", err
	}

	// получаем кол-во DC на кошельках
	sumWallets_, err := c.GetMap("SELECT currency_id, sum(amount) as sum_amount FROM wallets GROUP BY currency_id", "currency_id", "sum_amount")
	if err != nil {
		return "", err
	}
	sumWallets := make(map[int]float64)
	for currencyId, amount := range sumWallets_ {
		sumWallets[utils.StrToInt(currencyId)] = utils.StrToFloat64(amount)
	}

	// получаем кол-во TDC на обещанных суммах, плюсуем к тому, что на кошельках
	sumTdc, err := c.GetMap("SELECT currency_id, sum(tdc_amount) as sum_amount FROM promised_amount GROUP BY currency_id", "currency_id", "sum_amount")
	if err != nil {
		return "", err
	}

	for currencyId, amount := range sumTdc {
		currencyIdInt := utils.StrToInt(currencyId)
		if sumWallets[currencyIdInt] == 0 {
			sumWallets[currencyIdInt] = utils.StrToFloat64(amount)
		} else {
			sumWallets[currencyIdInt] += utils.StrToFloat64(amount)
		}
	}

	// получаем суммы обещанных сумм
	sumPromisedAmount, err := c.GetMap("SELECT currency_id, sum(amount) as sum_amount FROM promised_amount WHERE status = 'mining' AND del_block_id = 0 AND (cash_request_out_time = 0 OR cash_request_out_time > ?) GROUP BY currency_id", "currency_id", "sum_amount", time.Now().Unix()-c.Variables.Int64["cash_request_time"])
	if err != nil {
		return "", err
	}

	_, _, promisedAmountListGen, err := c.GetPromisedAmounts(c.SessUserId, c.Variables.Int64["cash_request_time"])

	calcTotal := utils.Round(100*math.Pow(1+currency_pct[72].MinerSec, 3600*24*30)-100, 0)

	// токен для запроса инфы с биржи
	var token, exchangeUrl string
	if c.SessRestricted == 0 {
		tokenAndUrl, err := c.OneRow(`SELECT token, e_host FROM ` + c.MyPrefix + `my_tokens LEFT JOIN miners_data ON miners_data.user_id = e_owner_id ORDER BY time DESC LIMIT 1`).String()
		if err != nil {
			return "", err
		}
		token = tokenAndUrl["token"]
		exchangeUrl = tokenAndUrl["e_host"]
	}

	myChatName := utils.Int64ToStr(c.SessUserId)
	// возможно у отпарвителя есть ник
	name, err := c.Single(`SELECT name FROM users WHERE user_id = ?`, c.SessUserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(name) > 0 {
		myChatName = name
	}

	// получим топ 5 бирж
	topExMap := make(map[int64]*topEx)
	var q string
	if c.ConfigIni["db_type"] == "postgresql" {
		//q = "SELECT DISTINCT e_owner_id, e_host, count(votes_exchange.user_id), result from votes_exchange LEFT JOIN miners_data ON votes_exchange.e_owner_id = miners_data.user_id WHERE e_host != '' GROUP BY e_owner_id, result, e_host"
		q = "SELECT DISTINCT e_owner_id, e_host, count(votes_exchange.user_id), result from miners_data  LEFT JOIN votes_exchange ON votes_exchange.e_owner_id = miners_data.user_id WHERE e_host != '' AND result >= 0 GROUP BY e_owner_id, result, e_host"
	} else {
		//q = "SELECT e_owner_id, e_host, count(votes_exchange.user_id) as count, result FROM miners_data LEFT JOIN votes_exchange ON votes_exchange.e_owner_id = miners_data.user_id WHERE and e_host != '' GROUP BY votes_exchange.e_owner_id, votes_exchange.result LIMIT 10"
		q = "SELECT e_owner_id, e_host, count(votes_exchange.user_id) as count, result FROM miners_data LEFT JOIN votes_exchange ON votes_exchange.e_owner_id = miners_data.user_id WHERE e_host != '' AND result >= 0 GROUP BY votes_exchange.e_owner_id, votes_exchange.result LIMIT 10"
	}
	rows, err := c.Query(q)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var user_id, count, result int64
		var e_host []byte
		err = rows.Scan(&user_id, &e_host, &count, &result)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if topExMap[user_id] == nil {
			topExMap[user_id] = new(topEx)
		}
		//if len(topExMap[user_id].Host) == 0 {
		//	topExMap[user_id] = new(topEx)
		if result == 0 {
			topExMap[user_id].Vote1 = count
		} else {
			topExMap[user_id].Vote1 = count
		}
		topExMap[user_id].Host = string(e_host)
		topExMap[user_id].UserId = user_id
		//}
	}

	// майнер ли я?
	miner_, err := c.Single(`SELECT miner_id FROM miners_data WHERE user_id = ?`, c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	var miner bool
	if miner_ > 0 {
		miner = true
	}

	refPhotos := make(map[int64][]string)

	// таблица обмена на наличные
	lastCashRequests, err := c.GetAll(`
			SELECT *
			FROM cash_requests
			ORDER BY id DESC
			LIMIT 5`, 5)
	for i := 0; i < len(lastCashRequests); i++ {
		if lastCashRequests[i]["del_block_id"] != "0" {
			lastCashRequests[i]["status"] = "reduction closed"
		} else if utils.Time()-utils.StrToInt64(lastCashRequests[i]["time"]) > c.Variables.Int64["cash_request_time"] && lastCashRequests[i]["status"] != "approved" {
			lastCashRequests[i]["status"] = "rejected"
		}
		t := time.Unix(utils.StrToInt64(lastCashRequests[i]["time"]), 0)
		lastCashRequests[i]["time"] = t.Format(c.TimeFormat)

		// ### from_user_id для фоток
		data, err := c.OneRow("SELECT * FROM miners_data WHERE user_id  =  ?", lastCashRequests[i]["from_user_id"]).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		// получим ID майнеров, у которых лежат фото нужного нам юзера
		minersIds := utils.GetMinersKeepers(data["photo_block_id"], data["photo_max_miner_id"], data["miners_keepers"], true)
		hosts, err := c.GetList("SELECT http_host FROM miners_data WHERE miner_id  IN (" + utils.JoinInts(minersIds, ",") + ")").String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		refPhotos[utils.StrToInt64(lastCashRequests[i]["from_user_id"])] = hosts

		// ### to_user_id для фоток
		data, err = c.OneRow("SELECT * FROM miners_data WHERE user_id  =  ?", lastCashRequests[i]["to_user_id"]).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		// получим ID майнеров, у которых лежат фото нужного нам юзера
		minersIds = utils.GetMinersKeepers(data["photo_block_id"], data["photo_max_miner_id"], data["miners_keepers"], true)
		hosts, err = c.GetList("SELECT http_host FROM miners_data WHERE miner_id  IN (" + utils.JoinInts(minersIds, ",") + ")").String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		refPhotos[utils.StrToInt64(lastCashRequests[i]["to_user_id"])] = hosts
	}

	var chart string
	// график обещанные суммы/монеты
	chartData, err := c.GetAll(`
			SELECT month, day, dc, promised_amount
			FROM stats
			WHERE currency_id = 72
			ORDER BY id DESC LIMIT 7`, 7)
	for i:=len(chartData)-1; i>=0; i-- {
		chart += `['`+chartData[i]["month"]+`/`+chartData[i]["day"]+`', `+utils.ClearNull(chartData[i]["promised_amount"], 0)+`, `+utils.ClearNull(chartData[i]["dc"], 0)+`],`
	}
	if len(chart) > 0 {
		chart = chart[:len(chart)-1]
	}
	
	charts := make([]ChartCur, 0)
	for _,icur := range [...]int64{72, 23} {
		chartData,_ := c.GetAll(`SELECT month, day, dc, promised_amount
				FROM stats	WHERE currency_id = ? ORDER BY id desc`, 30, icur )
		promised := make([]string, 0, 30);
		dc := make([]string, 0, 30);
		for _, val := range chartData {
			promised = append( promised, utils.ClearNull(val["promised_amount"], 0))
			dc = append( dc, utils.ClearNull(val["dc"], 0))
		}
		charts = append( charts, ChartCur{ currencyList[icur][1:], strings.Join( promised, `,` ), strings.Join( dc, `,` )})
	}
	DCTarget := consts.DCTarget[72]
	listBalance,_ := stat.TodayBalance( c.SessUserId )
	var statDays int
	if len(*listBalance) > 0 {
		statDays,_ = stat.GetHistoryBalance(listBalance, c.SessUserId)
	}
	//fmt.Println(`Stat`, statDays, err )
	var myDcTransactions []map[string]string
	if c.SessRestricted == 0 {
		myDcTransactions, err = c.GetAll("SELECT * FROM "+c.MyPrefix+"my_dc_transactions ORDER BY id DESC", 10)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		for id, data := range myDcTransactions {
			t := time.Unix(utils.StrToInt64(data["time"]), 0)
			timeFormatted := t.Format(c.TimeFormat)
			myDcTransactions[id]["timeFormatted"] = timeFormatted
			if utils.StrToInt64( data[`to_user_id`] ) ==  c.SessUserId {
				data[`sign`] = `+`
			} else {
				data[`sign`] = `-`
			}
		}
	}
	names := make(map[string]string)
	names["cash_request"] = c.Lang["cash"]
	names["from_mining_id"] = c.Lang["from_mining"]
	names["from_repaid"] = c.Lang["from_repaid_mining"]
	names["from_user"] = c.Lang["from_user"]
	names["node_commission"] = c.Lang["node_commission"]
	names["system_commission"] = c.Lang["system_commission"]
	names["referral"] = c.Lang["referral"]
	names["cf_project"] = "Crowd funding"
	names["cf_project_refund"] = "Crowd funding refund"

	TemplateStr, err := makeTemplate("home", "home", &homePage{
		DCTarget: DCTarget,
		Chart: 					chart,
		Charts:                 charts,
		Community:             c.Community,
		CountSignArr:          c.CountSignArr,
		CountSign:             c.CountSign,
		CalcTotal:             calcTotal,
		Admin:                 c.Admin,
		CurrencyPct:           currency_pct,
		SumWallets:            sumWallets,
		Wallets:               walletsByCurrency,
		PromisedAmountListGen: promisedAmountListGen,
		SessRestricted:        c.SessRestricted,
		SumPromisedAmount:     sumPromisedAmount,
		RandMiners:            randMiners,
		Points:                points,
		Assignments:           assignments,
		CurrencyList:          currencyList,
		ConfirmedBlockId:      confirmedBlockId,
		CashRequests:          cashRequests,
		LastCashRequests: lastCashRequests,
		ShowMap:               showMap,
		BlockId:               blockId,
		UserId:                c.SessUserId,
		PoolAdmin:             poolAdmin,
		Alert:                 c.Alert,
		MyNotice:              c.MyNotice,
		Lang:                  c.Lang,
		Title:                 c.Lang["geolocation"],
		ShowSignData:          c.ShowSignData,
		SignData:              "",
		MyChatName:            myChatName,
		IOS:                   utils.IOS(),
		Mobile:                utils.Mobile(),
		TopExMap:              topExMap,
		RefPhotos:                  refPhotos,

		ChatEnabled:           c.NodeConfig["chat_enabled"],
		Miner:                 miner,
		Token:                 token,
		ExchangeUrl:           exchangeUrl,
		ListBalance:           listBalance,
		StatDays:              statDays,
		MyDcTransactions:      myDcTransactions,
		Names:                 names })
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}

type topEx struct {
	Vote1  int64
	Vote0  int64
	Host   string
	UserId int64
}
