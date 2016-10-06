// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/astaxie/beego/config"
)

var (
	passMutex = sync.Mutex{}
	passUpd   = time.Now()
	passwords = make(map[string]bool)
	alphabet  = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
)

func genPass(length int) string {
	ret := make([]byte, length)
	for i := range ret {
		ret[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(ret)
}

func IsPassValid(pass, psw string) bool {
	passMutex.Lock()
	defer passMutex.Unlock()

	if len(passwords) == 0 || passUpd.Add(5*time.Minute).Before(time.Now()) {

		filename := *utils.Dir + `/passlist.txt`
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			out := make([]string, 1000)
			out[0] = pass
			for i := 1; i < 1000; i++ {
				out[i] = genPass(6)
			}
			ioutil.WriteFile(filename, []byte(strings.Join(out, "\r\n")), 0644)
		}
		if list, err := ioutil.ReadFile(filename); err == nil && len(list) > 0 {
			for key := range passwords {
				passwords[key] = false
			}
			out := strings.Split(string(list), "\r\n")
			for i := range out {
				plist := strings.SplitN(out[i], `,`, 2)
				if len(plist[0]) > 0 {
					passwords[plist[0]] = true
				}
			}
			passUpd = time.Now()
		}
	}
	return passwords[psw]
}

func Content(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("Content Recovered", fmt.Sprintf("%s: %s", e, debug.Stack()))
			fmt.Println("Content Recovered", fmt.Sprintf("%s: %s", e, debug.Stack()))
		}
	}()
	var err error

	w.Header().Set("Content-type", "text/html")

	sess, err := globalSessions.SessionStart(w, r)
	if err != nil {
		log.Error("%v", err)
	}
	defer sess.SessionRelease(w)
	sessWalletId := GetSessWalletId(sess)
	sessCitizenId := GetSessCitizenId(sess)
	sessStateId := GetSessInt64("state_id", sess)
	sessAddress := GetSessString(sess, "address")
	sessAccountId := GetSessInt64("account_id", sess)
	log.Debug("sessWalletId %v / sessCitizenId %v", sessWalletId, sessCitizenId)

	c := new(Controller)
	c.r = r
	c.w = w
	c.sess = sess
	c.SessWalletId = sessWalletId
	c.SessCitizenId = sessCitizenId
	c.SessStateId = sessStateId
	c.SessAddress = sessAddress

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
	stateName := ""
	if sessStateId > 0 {
		stateName, err = c.GetStateName(sessStateId)
		if err != nil {
			log.Error("%v", err)
		}
		c.StateName = stateName
		c.StateId = sessStateId
		c.StateIdStr = utils.Int64ToStr(sessStateId)
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

		// Инфа о последнем блоке
		blockData, err := c.DCDB.GetLastBlockData()
		if err != nil {
			log.Error("%v", err)
		}
		//время последнего блока
		lastBlockTime = blockData["lastBlockTime"]
		log.Debug("installProgress", installProgress, "configExists", configExists, "lastBlockTime", lastBlockTime)

		confirmedBlockId, err := c.GetConfirmedBlockId()
		if err != nil {
			log.Error("%v", err)
		}
		c.ConfirmedBlockId = confirmedBlockId

	}
	r.ParseForm()
	pageName := r.FormValue("page")

	tplName := r.FormValue("tpl_name")
	if len(tplName) == 0 {
		tplName = r.FormValue("controllerHTML")
		if len(tplName) == 0 {
			tplName = pageName
			if len(tplName) == 0 {
				tplName = "dashboardAnonym"
			}
		}
	}
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

	match, _ := regexp.MatchString("^(installStep[0-9_]+)|(blockExplorer)$", tplName)
	// CheckInputData - гарантирует, что tplName чист
	if tplName != "" && utils.CheckInputData(tplName, "tpl_name") && (sessWalletId > 0 || sessCitizenId > 0 || len(sessAddress) > 0 || match) {
		tplName = tplName
	} else if dbInit && installProgress == "complete" && len(configExists) == 0 {
		// первый запуск, еще не загружен блокчейн
		tplName = "updatingBlockchain"
	} else if dbInit && installProgress == "complete" && (sessWalletId > 0 || sessCitizenId > 0 || len(sessAddress) > 0) {
		tplName = "dashboardAnonym"
	} else if dbInit && installProgress == "complete" {
		if tplName != "loginECDSA" {
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
		tplName = "updatingBlockchain"
	}
	log.Debug("lastBlockTime %v / utils.Time() %v / wTime %v", lastBlockTime, utils.Time(), wTime)

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
	if (sessWalletId > 0 || sessCitizenId > 0) && dbInit && installProgress == "complete" {
		countSign = 1
		log.Debug("userId: %d", userId)
		var pk map[string]string
		if sessWalletId > 0 {
			pk, err = c.OneRow("SELECT hex(public_key_1) as public_key_1, hex(public_key_2) as public_key_2 FROM dlt_wallets WHERE wallet_id = ?", userId).String()
		} else {
			pk, err = c.OneRow(`SELECT hex(public_key_1) as public_key_1, hex(public_key_2) as public_key_2 FROM `+c.StateIdStr+`_citizens WHERE citizen_id = ?`, userId).String()
		}
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
	}

	log.Debug("countSign: %v", countSign)
	var CountSignArr []int
	for i := 0; i < countSign; i++ {
		CountSignArr = append(CountSignArr, i)
	}
	c.CountSign = countSign
	c.CountSignArr = CountSignArr

	if tplName == "" {
		tplName = "login"
	}

	log.Debug("tplName::", tplName, sessCitizenId, sessWalletId, installProgress)

	fmt.Println("tplName::", tplName, sessCitizenId, sessWalletId, sessAddress)
	controller := r.FormValue("controllerHTML")
	if val, ok := configIni[`psw`]; ok && ((tplName != `login` && tplName != `loginECDSA`) || len(controller) > 0) {
		if psw, err := r.Cookie(`psw`); err != nil || !IsPassValid(val, psw.Value) {
			if err == nil {
				cookie := http.Cookie{Name: "psw", Value: ``, Expires: time.Now().AddDate(0, 0, -1)}
				http.SetCookie(w, &cookie)
			}
			if controller == `menu` || tplName == `menu` || tplName == `ModalAnonym` {
				w.Write([]byte{})
				return
			}
			c.Logout()
			controller = `psw`
			pageName = ``
			tplName = ``
		}
	}

	if len(controller) > 0 {
		fmt.Println(`Controller HTML`, controller)
		log.Debug("controller:", controller)

		funcMap := template.FuncMap{
			"noescape": func(s string) template.HTML {
				return template.HTML(s)
			},
		}
		data, err := static.Asset("static/" + controller + ".html")
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
	if len(pageName) > 0 && isPage(pageName, TPage) {
		c.Data = &CommonPage{
			Address:      c.SessAddress,
			WalletId:     c.SessWalletId,
			CitizenId:    c.SessCitizenId,
			StateId:      c.SessStateId,
			StateName:    stateName,
			CountSignArr: []byte{1}, // !!! Добавить вычисление
		}
		w.Write([]byte(CallPage(c, pageName)))
		return
	}
	if ok, _ := regexp.MatchString(`^(?i)listOfTables|editStateParameters|editColumn|contracts|newContract|editContract|editMenu|newMenu|newPage|editPage|editMenu|newColumn|editTable|showTable|stateTable|newState|tableList|newTable|stateLaws|stateSmartLaws|changeStateParameters|stateParameters|blockGeneration|LoginECDSA|AnonymMoneyTransfer|ModalAnonym|DashBoardAnonym|Transactions|NotificationList|Map|PromisedAmountRestricted|PromisedAmountRestrictedList|upgradeUser|miningSn|changePool|delPoolUser|delAutoPayment|newAutoPayment|autoPayments|holidaysList|adminVariables|adminSpots|exchangeAdmin|exchangeSupport|exchangeUser|votesExchange|chat|firstSelect|PoolAdminLogin|CfPagePreview|CfCatalog|AddCfProjectData|CfProjectChangeCategory|NewCfProject|MyCfProjects|DelCfProject|DelCfFunding|CfStart|PoolAdminControl|Credits|Home|WalletsList|Information|Notifications|Interface|MiningMenu|Upgrade5|NodeConfigControl|Upgrade7|Upgrade6|Upgrade5|Upgrade4|Upgrade3|Upgrade2|Upgrade1|Upgrade0|StatisticVoting|ProgressBar|MiningPromisedAmount|CurrencyExchangeDelete|CurrencyExchange|ChangeCreditor|ChangeCommission|CashRequestOut|ArbitrationSeller|ArbitrationBuyer|ArbitrationArbitrator|Arbitration|InstallStep2|InstallStep1|InstallStep0|DbInfo|ChangeHost|Assignments|NewUser|NewPhoto|Voting|VoteForMe|RepaymentCredit|PromisedAmountList|PromisedAmountActualization|NewPromisedAmount|Login|ForRepaidFix|DelPromisedAmount|DelCredit|ChangePromisedAmount|ChangePrimaryKey|ChangeNodeKey|ChangeAvatar|BugReporting|Abuse|UpgradeResend|UpdatingBlockchain|Statistic|RewritePrimaryKey|RestoringAccess|PoolTechWorks|Points|NewHolidays|NewCredit|MoneyBackRequest|MoneyBack|ChangeMoneyBack|ChangeKeyRequest|ChangeKeyClose|ChangeGeolocation|ChangeCountryRace|ChangeArbitratorConditions|CashRequestIn|BlockExplorer$`, tplName); !ok {
		w.Write([]byte("Access denied 0"))
	} else if len(tplName) > 0 && (sessCitizenId > 0 || sessWalletId > 0 || len(sessAddress) > 0) && installProgress == "complete" {

		if tplName == "login" {
			tplName = "dashboard_anonym"
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

		if dbInit {
			// Если у юзера только 1 праймари ключ, то выдавать форму, где показываются данные для подписи и форма ввода подписи не нужно.
			// Только если он сам не захочет, указав это в my_table
			showSignData := false
			if showSignData || countSign > 1 {
				c.ShowSignData = true
			} else {
				c.ShowSignData = false
			}
		}

		if dbInit && tplName != "updatingBlockchain" {
			html, err := CallController(c, "AlertMessage")
			if err != nil {
				log.Error("%v", err)
			}
			w.Write([]byte(html))
		}
		w.Write([]byte("<input type='hidden' id='tpl_name' value='" + tplName + "'>"))

		log.Debug("tplName==", tplName)

		// подсвечиваем красным номер блока, если идет процесс обновления
		var blockJs string
		blockId, err := c.GetBlockId()
		if err != nil {
			log.Error("%v", err)
		}
		blockJs = "$('#block_id').html(" + utils.Int64ToStr(blockId) + ");$('#block_id').css('color', '#428BCA');"

		w.Write([]byte(`<script>
								$( document ).ready(function() {
								` + blockJs + `
								});
								</script>`))
		skipRestrictedUsers := []string{"cashRequestIn", "cashRequestOut", "upgrade", "notifications"}

		if c.StateId > 0 && (tplName == "dashboard_anonym" || tplName == "home") {
			tpl, err := utils.CreateHtmlFromTemplate("dashboard_default", sessCitizenId, sessAccountId, sessStateId)
			if err != nil {
				log.Error("%v", err)
				return
			}
			w.Write([]byte(tpl))
			return
		}

		// тем, кто не зареган на пуле не выдаем некоторые страницы
		if !utils.InSliceString(tplName, skipRestrictedUsers) {
			// вызываем контроллер в зависимости от шаблона
			html, err := CallController(c, tplName)
			if err != nil {
				log.Error("%v", err)
			}
			w.Write([]byte(html))
		}
	} else if len(tplName) > 0 {
		if tplName == "login" {
			tplName = "LoginECDSA"
		}

		log.Debug("tplName", tplName)
		html := ""
		if ok, _ := regexp.MatchString(`^(?i)LoginECDSA|blockExplorer|CfCatalog|CfPagePreview|CfStart|Check_sign|CheckNode|GetBlock|GetMinerData|GetMinerDataMap|GetSellerData|Index|IndexCf|InstallStep0|InstallStep1|InstallStep2|Login|SignLogin|SynchronizationBlockchain|UpdatingBlockchain|Menu$`, tplName); !ok && c.SessCitizenId <= 0 && c.SessWalletId <= 0 && len(c.SessAddress) == 0 {
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
		html, err := CallController(c, "LoginECDSA")
		if err != nil {
			log.Error("%v", err)
		}
		w.Write([]byte(html))
	}

	//sess.Set("username", 11111)

}
