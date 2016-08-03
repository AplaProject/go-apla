package controllers

import (
	"bytes"
	"encoding/json"
	"github.com/astaxie/beego/config"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"net/http"
	"os"
	"regexp"
	"fmt"
	"html/template"
	"github.com/DayLightProject/go-daylight/packages/static"
	"runtime/debug"
)

func Content(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("Content Recovered", fmt.Sprintf("%s: %s", e, debug.Stack()))
			fmt.Println("Content Recovered", fmt.Sprintf("%s: %s", e, debug.Stack()))
		}
	}()
	var err error

	w.Header().Set("Content-type", "text/html")

	// чтобы в чат не вставлялись старые сообщения после новых
	utils.ChatMinSignTime = 0

	sess, err := globalSessions.SessionStart(w, r)
	if err != nil {
		log.Error("%v", err)
	}
	defer sess.SessionRelease(w)
	sessUserId := GetSessUserId(sess)
	sessRestricted := GetSessRestricted(sess)
	sessPublicKey := GetSessPublicKey(sess)
	sessAdmin := GetSessAdmin(sess)
	log.Debug("sessUserId", sessUserId)
	log.Debug("sessRestricted", sessRestricted)
	log.Debug("sessPublicKey", sessPublicKey)
	log.Debug("user_id: %v", sess.Get("user_id"))

	c := new(Controller)
	c.r = r
	c.w = w
	c.sess = sess
	c.SessRestricted = sessRestricted
	c.SessUserId = sessUserId
	if sessAdmin == 1 {
		c.Admin = true
	}
	c.ContentInc = true

	var installProgress, configExists string
	var lastBlockTime int64

	dbInit := false
	if len(configIni["db_user"]) > 0 || (configIni["db_type"] == "sqlite") {
		dbInit = true
	}

	if dbInit {
		var err error
		//c.DCDB, err = utils.NewDbConnect(configIni)
		c.DCDB = utils.DB
		if c.DCDB.DB == nil {
			log.Error("utils.DB == nil")
			dbInit = false
		}
		if dbInit {
			// отсутвие таблы выдаст ошибку, значит процесс инсталяции еще не пройден и надо выдать 0-й шаг
			_, err = c.DCDB.Single("SELECT progress FROM install").String()
			if err != nil {
				log.Error("%v", err)
				dbInit = false
			}
		}
	}

	c.dbInit = dbInit

	if dbInit {
		var err error
		installProgress, err = c.DCDB.Single("SELECT progress FROM install").String()
		if err != nil {
			log.Error("%v", err)
		}
		configExists, err = c.DCDB.Single("SELECT first_load_blockchain_url FROM config").String()
		if err != nil {
			log.Error("%v", err)
		}

		c.Variables, err = c.GetAllVariables()

		// Инфа о последнем блоке
		blockData, err := c.DCDB.GetLastBlockData()
		if err != nil {
			log.Error("%v", err)
		}
		//время последнего блока
		lastBlockTime = blockData["lastBlockTime"]
		log.Debug("installProgress", installProgress, "configExists", configExists, "lastBlockTime", lastBlockTime)

		currencyList, err := c.GetCurrencyList(false)
		if err != nil {
			log.Error("%v", err)
		}
		c.CurrencyList = currencyList

		confirmedBlockId, err := c.GetConfirmedBlockId()
		if err != nil {
			log.Error("%v", err)
		}
		c.ConfirmedBlockId = confirmedBlockId


	}
	r.ParseForm()
	tplName := r.FormValue("tpl_name")

	c.Parameters, err = c.GetParameters()
	log.Debug("parameters=", c.Parameters)

	log.Debug("tpl_name=", tplName)

	// если в параметрах пришел язык, то установим его
	newLang := utils.StrToInt(c.Parameters["lang"])
	if newLang > 0 {
		log.Debug("newLang", newLang)
		SetLang(w, r, newLang)
	}
	// уведомления
	//if utils.CheckInputData(parameters["alert"], "alert") {
	c.Alert = c.Parameters["alert"]
	//}

	lang := GetLang(w, r, c.Parameters)
	log.Debug("lang", lang)

	c.Lang = globalLangReadOnly[lang]
	c.LangInt = int64(lang)
	if lang == 42 {
		c.TimeFormat = "2006-01-02 15:04:05"
	} else {
		c.TimeFormat = "2006-02-01 15:04:05"
	}

	c.Periods = map[int64]string{86400: "1 " + c.Lang["day"], 604800: "1 " + c.Lang["week"], 31536000: "1 " + c.Lang["year"], 2592000: "1 " + c.Lang["month"], 1209600: "2 " + c.Lang["weeks"]}

	c.Races = map[int64]string{1: c.Lang["race_1"], 2: c.Lang["race_2"], 3: c.Lang["race_3"]}
	var status string
	var communityUsers []int64
	if dbInit {
		communityUsers, err = c.DCDB.GetCommunityUsers()
		if err != nil {
			log.Error("%v", err)
		}
		c.CommunityUsers = communityUsers
		if len(communityUsers) == 0 {
			c.MyPrefix = ""
		} else {
			c.MyPrefix = utils.Int64ToStr(sessUserId) + "_"
			c.Community = true
		}
		log.Debug("c.MyPrefix %s", c.MyPrefix)
		// нужна мин. комиссия на пуле для перевода монет
		config, err := c.GetNodeConfig()
		if err != nil {
			log.Error("%v", err)
		}
		configCommission_ := make(map[string][]float64)
		if len(config["commission"]) > 0 {
			err = json.Unmarshal([]byte(config["commission"]), &configCommission_)
			if err != nil {
				log.Error("%v", err)
			}
		}
		configCommission := make(map[int64][]float64)
		for k, v := range configCommission_ {
			configCommission[utils.StrToInt64(k)] = v
		}
		c.NodeConfig = config
		c.ConfigCommission = configCommission

		c.NodeAdmin, err = c.NodeAdminAccess(c.SessUserId, c.SessRestricted)
		if err != nil {
			log.Error("%v", err)
		}

		status, err = c.DCDB.Single("SELECT status FROM " + c.MyPrefix + "my_table").String()
		if err != nil {
			log.Error("%v", err)
		}
	}
	log.Debug("dbInit", dbInit)

	setupPassword := c.NodeConfig["setup_password"]
	match, _ := regexp.MatchString("^(installStep[0-9_]+)|(blockExplorer)$", tplName)
	// CheckInputData - гарантирует, что tplName чист
	if tplName != "" && utils.CheckInputData(tplName, "tpl_name") && (sessUserId > 0 || match) {
		tplName = tplName
	} else if dbInit && installProgress == "complete" && len(configExists) == 0 {
		// первый запуск, еще не загружен блокчейн
		tplName = "updatingBlockchain"
	} else if dbInit && installProgress == "complete" && sessUserId > 0 {
		if status == "waiting_set_new_key" {
			tplName = "setPassword"
		} else if status == "waiting_accept_new_key" {
			tplName = "waitingAcceptNewKey"
		}
	} else if dbInit && installProgress == "complete" && !c.Community && sessUserId == 0 && status == "waiting_set_new_key" && setupPassword != "" {
		tplName = "setupPassword"
	} else if dbInit && installProgress == "complete" && sessUserId == 0 && status == "waiting_accept_new_key" {
		tplName = "waitingAcceptNewKey"
	} else if dbInit && installProgress == "complete" {
		if tplName != "setPassword" {
			tplName = "login"
		}
	} else {
		tplName = "installStep0" // самый первый запуск
	}
	log.Debug("dbInit", dbInit, "installProgress", installProgress, "configExists", configExists)
	log.Debug("tplName>>>>>>>>>>>>>>>>>>>>>>", tplName)

	// идет загрузка блокчейна
	wTime := int64(2)
	if configIni != nil && configIni["test_mode"] == "1" {
		wTime = 2 * 365 * 86400
		log.Debug("%v", wTime)
		log.Debug("%v", lastBlockTime)
	}
	if dbInit && tplName != "installStep0" && (utils.Time()-lastBlockTime > 3600*wTime) && len(configExists) > 0 {
		if len(communityUsers) > 0 {
			// исключение - админ пула
			poolAdminUserId, err := c.DCDB.Single("SELECT pool_admin_user_id FROM config").String()
			if err != nil {
				log.Error("%v", err)
			}
			if sessUserId != utils.StrToInt64(poolAdminUserId) {
				tplName = "updatingBlockchain"
			}
		} else {
			tplName = "updatingBlockchain"
		}
	}

	if tplName == "installStep0" {
		log.Debug("ConfigInit monitor")
		if _, err := os.Stat(*utils.Dir + "/config.ini"); err == nil {

			configIni_, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			configIni, err = configIni_.GetSection("default")
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			if len(configIni["db_type"]) > 0 {
				tplName = "updatingBlockchain"
			}
		}
	}

	log.Debug("tplName2=", tplName)

	// кол-во ключей=подписей у юзера
	var countSign int
	var userId int64
	//	var myUserId int64
	if sessUserId > 0 && dbInit && installProgress == "complete" {
		userId = sessUserId
		//myUserId = sessUserId
		countSign = 1
		log.Debug("userId: %d", userId)
		pk, err := c.OneRow("SELECT hex(public_key_1) as public_key_1, hex(public_key_2) as public_key_2 FROM users WHERE user_id = ?", userId).String()
		if err != nil {
			log.Error("%v", err)
		}
		log.Debug("pk: %v", pk)
		if len(pk["public_key_1"]) > 0 {
			log.Debug("public_key_1: %x", pk["public_key_1"])
			countSign = 2
		}
		if len(pk["public_key_2"]) > 0 {
			log.Debug("public_key_2: %x", pk["public_key_2"])
			countSign = 3
		}
	} else {
		userId = 0
		//myUserId = 0
	}

	log.Debug("countSign: %v", countSign)
	c.UserId = userId
	var CountSignArr []int
	for i := 0; i < countSign; i++ {
		CountSignArr = append(CountSignArr, i)
	}
	c.CountSign = countSign
	c.CountSignArr = CountSignArr

	if tplName == "" {
		tplName = "login"
	}

	log.Debug("tplName::", tplName, sessUserId, installProgress)

	controller := r.FormValue("controllerHTML")
	if len(controller) > 0 {

		log.Debug("controller:", controller)

		funcMap := template.FuncMap{
			"noescape": func(s string) template.HTML {
				return template.HTML(s)
			},
		}
		data, err := static.Asset("static/"+controller+".html")
		t := template.New("template").Funcs(funcMap)
		t, err = t.Parse(string(data))
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		}
		
		b := new(bytes.Buffer)
		err = t.Execute(b, c)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		}
		w.Write(b.Bytes())
		return
	}

	if ok, _ := regexp.MatchString(`^(?i)Transactions|NotificationList|Map|PromisedAmountRestricted|PromisedAmountRestrictedList|upgradeUser|miningSn|changePool|delPoolUser|delAutoPayment|newAutoPayment|autoPayments|holidaysList|adminVariables|adminSpots|exchangeAdmin|exchangeSupport|exchangeUser|votesExchange|chat|firstSelect|PoolAdminLogin|setupPassword|waitingAcceptNewKey|SetPassword|CfPagePreview|CfCatalog|AddCfProjectData|CfProjectChangeCategory|NewCfProject|MyCfProjects|DelCfProject|DelCfFunding|CfStart|PoolAdminControl|Credits|Home|WalletsList|Information|Notifications|Interface|MiningMenu|Upgrade5|NodeConfigControl|Upgrade7|Upgrade6|Upgrade5|Upgrade4|Upgrade3|Upgrade2|Upgrade1|Upgrade0|StatisticVoting|ProgressBar|MiningPromisedAmount|CurrencyExchangeDelete|CurrencyExchange|ChangeCreditor|ChangeCommission|CashRequestOut|ArbitrationSeller|ArbitrationBuyer|ArbitrationArbitrator|Arbitration|InstallStep2|InstallStep1|InstallStep0|DbInfo|ChangeHost|Assignments|NewUser|NewPhoto|Voting|VoteForMe|RepaymentCredit|PromisedAmountList|PromisedAmountActualization|NewPromisedAmount|Login|ForRepaidFix|DelPromisedAmount|DelCredit|ChangePromisedAmount|ChangePrimaryKey|ChangeNodeKey|ChangeAvatar|BugReporting|Abuse|UpgradeResend|UpdatingBlockchain|Statistic|RewritePrimaryKey|RestoringAccess|PoolTechWorks|Points|NewHolidays|NewCredit|MoneyBackRequest|MoneyBack|ChangeMoneyBack|ChangeKeyRequest|ChangeKeyClose|ChangeGeolocation|ChangeCountryRace|ChangeArbitratorConditions|CashRequestIn|BlockExplorer$`, tplName); !ok {
		w.Write([]byte("Access denied 0"))
	} else if len(tplName) > 0 && sessUserId > 0 && installProgress == "complete" {
		// если ключ юзера изменился, то выбрасываем его
		userPublicKey, err := c.DCDB.GetUserPublicKey(userId)
		if err != nil {
			log.Error("%v", err)
		}
		// но возможно у юзера включено сохранение приватного ключа
		// тогда, чтобы не получилось зацикливания, нужно проверить и my_keys
		myPrivateKey, err := c.GetMyPrivateKey(c.MyPrefix)
		if err != nil {
			log.Error("%v", err)
		}
		myPublicKey, err := c.GetMyPublicKey(c.MyPrefix)
		if err != nil {
			log.Error("%v", err)
		}
		/* !!! Зачем эта проверка нужна?
			Если sessUserId > 0 && installProgress == "complete", то в users точно должно что-то быть
		countUsers, err := c.Single(`SELECT count(*) FROM users`).Int64()
		if err != nil {
			log.Error("%v", err)
		}*/
		if (string(utils.BinToHex(userPublicKey)) != sessPublicKey && len(myPrivateKey) == 0) || (/*countUsers > 0 &&*/ len(myPrivateKey) > 0 && !bytes.Equal(myPublicKey, []byte(userPublicKey))) {
			log.Debug("userPublicKey!=sessPublicKey %s!=%s / userId: %d", utils.BinToHex(userPublicKey), sessPublicKey, userId)
			log.Debug("len(myPrivateKey) = %d  && %x!=%x", len(myPrivateKey), string(myPublicKey), userPublicKey)
			c.Logout()
			if len(userPublicKey) > 0 {
				w.Write([]byte("<script language=\"javascript\">window.location.href = \"/\"</script>If you are not redirected automatically, follow the <a href=\"/\">/</a>"))
				return
			} 
		}

		if tplName == "login" {
			tplName = "home"
		}

/*		if tplName == "home" && c.Parameters["first_select"] != "1" {
			data, err := c.OneRow(`SELECT first_select, miner_id from ` + c.MyPrefix + `my_table`).Int64()
			if err != nil {
				log.Error("%v", err)
			}
			if data["first_select"] == 0 && data["miner_id"] == 0 && c.SessRestricted == 0 {
				tplName = "firstSelect"
			}
		} */
		c.TplName = tplName

		log.Debug("communityUsers:", communityUsers)
		if dbInit && len(communityUsers) > 0 {
			poolAdminUserId, err := c.GetPoolAdminUserId()
			if err != nil {
				log.Error("%v", err)
			}
			c.PoolAdminUserId = poolAdminUserId
			if c.SessUserId == poolAdminUserId {
				c.PoolAdmin = true
			}
		} else {
			c.PoolAdmin = true
		}

		if dbInit {
			// проверим, не идут ли тех. работы на пуле
			config, err := c.DCDB.OneRow("SELECT pool_admin_user_id, pool_tech_works FROM config").String()
			if err != nil {
				log.Error("%v", err)
			}
			if len(config["pool_admin_user_id"]) > 0 && utils.StrToInt64(config["pool_admin_user_id"]) != sessUserId && config["pool_tech_works"] == "1" && c.Community {
				tplName = "login"
			}
			// Если у юзера только 1 праймари ключ, то выдавать форму, где показываются данные для подписи и форма ввода подписи не нужно.
			// Только если он сам не захочет, указав это в my_table
			showSignData := false
			if sessRestricted == 0 { // у незареганных в пуле юзеров нет MyPrefix, поэтому сохранять значение show_sign_data им негде
				showSignData_, err := c.DCDB.Single("SELECT show_sign_data FROM " + c.MyPrefix + "my_table").String()
				if err != nil {
					log.Error("%v", err)
				}
				if showSignData_ == "1" {
					showSignData = true
				} else {
					showSignData = false
				}
			}
			if showSignData || countSign > 1 {
				c.ShowSignData = true
			} else {
				c.ShowSignData = false
			}
		}

		// писать в чат можно и при апдейте блокчейна
		if r.FormValue("tpl_name") == "chat" && tplName == "updatingBlockchain" {
			tplName = "chat"
		}

		if dbInit && tplName != "updatingBlockchain" && tplName != "setPassword" && tplName != "waitingAcceptNewKey" {
			html, err := CallController(c, "AlertMessage")
			if err != nil {
				log.Error("%v", err)
			}
			w.Write([]byte(html))
		}
		w.Write([]byte("<input type='hidden' id='tpl_name' value='" + tplName + "'>"))

		myNotice, err := c.DCDB.GetMyNoticeData(sessRestricted, sessUserId, c.MyPrefix, globalLangReadOnly[lang])
		if err != nil {
			log.Error("%v", err)
		}
		c.MyNotice = myNotice

		log.Debug("tplName==", tplName)

		// подсвечиваем красным номер блока, если идет процесс обновления
		var blockJs string
		blockId, err := c.GetBlockId()
		if err != nil {
			log.Error("%v", err)
		}
		if myNotice["main_status_complete"] != "1" {
			blockJs = "$('#block_id').html(" + utils.Int64ToStr(blockId) + ");$('#block_id').css('color', '#ff0000');"
		} else {
			blockJs = "$('#block_id').html(" + utils.Int64ToStr(blockId) + ");$('#block_id').css('color', '#428BCA');"
		}
		w.Write([]byte(`<script>
								$( document ).ready(function() {
								` + blockJs + `
								});
								</script>`))
		skipRestrictedUsers := []string{"cashRequestIn", "cashRequestOut", "upgrade", "notifications"}
		// тем, кто не зареган на пуле не выдаем некоторые страницы
		if sessRestricted == 0 || !utils.InSliceString(tplName, skipRestrictedUsers) {
			// вызываем контроллер в зависимости от шаблона
			html, err := CallController(c, tplName)
			if err != nil {
				log.Error("%v", err)
			}
			w.Write([]byte(html))
		}
	} else if tplName == "setPassword" {
		html, err := CallController(c, tplName)
		if err != nil {
			log.Error("%v", err)
		}
		w.Write([]byte(html))
	} else if len(tplName) > 0 {
		log.Debug("tplName", tplName)
		html := ""
		if ok, _ := regexp.MatchString(`^(?i)blockExplorer|waitingAcceptNewKey|SetupPassword|CfCatalog|CfPagePreview|CfStart|Check_sign|CheckNode|GetBlock|GetMinerData|GetMinerDataMap|GetSellerData|Index|IndexCf|InstallStep0|InstallStep1|InstallStep2|Login|SignLogin|SynchronizationBlockchain|UpdatingBlockchain|Menu$`, tplName); !ok && c.SessUserId <= 0 {
			html = "Access denied 1"
		} else {
			// если сессия обнулилась в процессе навигации по админке, то вместо login шлем на /, чтобы очистилось меню
			if len(r.FormValue("tpl_name")) > 0 && tplName == "login" {
				log.Debug("window.location.href = /")
				w.Write([]byte("<script language=\"javascript\">window.location.href = \"/\"</script>If you are not redirected automatically, follow the <a href=\"/\">/</a>"))
				return
			}
			// вызываем контроллер в зависимости от шаблона
			html, err = CallController(c, tplName)
			if err != nil {
				log.Error("%v", err)
			}
		}
		w.Write([]byte(html))
	} else {
		html, err := CallController(c, "login")
		if err != nil {
			log.Error("%v", err)
		}
		w.Write([]byte(html))
	}
	//sess.Set("username", 11111)

}
