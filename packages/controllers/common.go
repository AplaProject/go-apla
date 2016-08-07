package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/session"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/op/go-logging"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var log = logging.MustGetLogger("controllers")

type Controller struct {
	dbInit bool
	*utils.DCDB
	r                *http.Request
	w                http.ResponseWriter
	sess             session.SessionStore
	Lang             map[string]string
	TplName          string
	LangInt          int64
	Navigate         string
	ContentInc       bool
	Periods          map[int64]string
	Community        bool
	CommunityUsers   []int64
	ShowSignData     bool
	MyPrefix         string
	Alert            string
	UserId           int64
	Admin            bool
	SessCitizenId    int64
	SessWalletId     int64
	MyNotice         map[string]string
	Parameters       map[string]string
	Variables        *utils.Variables
	CountSign        int
	CountSignArr     []int
	TimeFormat       string
	PoolAdmin        bool
	PoolAdminUserId  int64
	NodeAdmin        bool
	NodeConfig       map[string]string
	ConfigCommission map[int64][]float64
	CurrencyList     map[int64]string
	CurrencyListCf   map[int64]string
	PaymentSystems   map[string]string
	ConfirmedBlockId int64
	MinerId          int64
	Races            map[int64]string
	EConfig          map[string]string
	ECommission      float64
	EURL             string
}

var (
	configIni      map[string]string
	globalSessions *session.Manager
	// в гоурутинах используется только для чтения
	globalLangReadOnly map[int]map[string]string
)

func SessInit() {
	var err error
	/*path := *utils.Dir + `/tmp`
	if runtime.GOOS == "windows" {
		path = "tmp"
	}
	globalSessions, err = session.NewManager("file", `{"cookieName":"gosessionid","gclifetime":864000,"ProviderConfig":"`+path+`"}`)*/
	globalSessions, err = session.NewManager("memory", `{"cookieName":"gosessionid","gclifetime":864000}`)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
	}
	go globalSessions.GC()
}

func ConfigInit() {

	// мониторим config.ini на наличие изменений
	go func() {
		for {
			log.Debug("ConfigInit monitor")
			if _, err := os.Stat(*utils.Dir + "/config.ini"); os.IsNotExist(err) {
				utils.Sleep(1)
				continue
			}
			configIni_, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			configIni, err = configIni_.GetSection("default")
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			if len(configIni["db_type"]) > 0 {
				break
			}
			utils.Sleep(3)
		}
	}()
	globalLangReadOnly = make(map[int]map[string]string)
	for _, v := range consts.LangMap {
		data, err := static.Asset(fmt.Sprintf("static/lang/%d.ini", v))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
		iniconf_, err := config.NewConfigData("ini", []byte(data))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
		//fmt.Println(iniconf_)
		iniconf, err := iniconf_.GetSection("default")
		globalLangReadOnly[v] = make(map[string]string)
		globalLangReadOnly[v] = iniconf
	}
}

func init() {
	flag.Parse()
}

func CallController(c *Controller, name string) (string, error) {
	// имя экспортируемого метода должно начинаться с заглавной буквы
	a := []rune(name)
	a[0] = unicode.ToUpper(a[0])
	name = string(a)
	log.Debug("Controller %v", name)
	html, err := CallMethod(c, name)
	if err != nil {
		log.Error("err: %v / Controller: %v", err, name)
		html = fmt.Sprintf(`{"error":%q}`, err)
		log.Debug("%v", html)
	}
	return html, err
}

func CallMethod(i interface{}, methodName string) (string, error) {
	var ptr reflect.Value
	var value reflect.Value
	var finalMethod reflect.Value

	value = reflect.ValueOf(i)

	// if we start with a pointer, we need to get value pointed to
	// if we start with a value, we need to get a pointer to that value
	if value.Type().Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(value)
	}

	// check for method on value
	method := value.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}
	// check for method on pointer
	method = ptr.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}

	if finalMethod.IsValid() {
		x := finalMethod.Call([]reflect.Value{})
		err_, found := x[1].Interface().(error)
		var err error
		if found {
			err = err_
		} else {
			err = nil
		}
		return x[0].Interface().(string), err
	}

	// return or panic, method not found of either type
	return "", fmt.Errorf("method not found")
}

func GetSessEUserId(sess session.SessionStore) int64 {
	sessUserId := sess.Get("e_user_id")
	log.Debug("sessUserId: %v", sessUserId)
	switch sessUserId.(type) {
	case int64:
		return sessUserId.(int64)
	case int:
		return int64(sessUserId.(int))
	case string:
		return utils.StrToInt64(sessUserId.(string))
	default:
		return 0
	}
	return 0
}
func GetSessWalletId(sess session.SessionStore) int64 {
	sessUserId := sess.Get("wallet_id")
	log.Debug("sessUserId: %v", sessUserId)
	switch sessUserId.(type) {
	case int64:
		return sessUserId.(int64)
	case int:
		return int64(sessUserId.(int))
	case string:
		return utils.StrToInt64(sessUserId.(string))
	default:
		return 0
	}
	return 0
}


func GetSessCitizenId(sess session.SessionStore) int64 {
	sessUserId := sess.Get("citizen_id")
	log.Debug("sessUserId: %v", sessUserId)
	switch sessUserId.(type) {
		case int64:
		return sessUserId.(int64)
		case int:
		return int64(sessUserId.(int))
		case string:
		return utils.StrToInt64(sessUserId.(string))
		default:
		return 0
	}
	return 0
}

func GetSessAdmin(sess session.SessionStore) int64 {
	admin := sess.Get("admin")
	switch admin.(type) {
	case int64:
		return admin.(int64)
	default:
		return 0
	}
	return 0
}

func DelSessResctricted(sess session.SessionStore) {
	sess.Delete("restricted")
}

func GetSessInt64(sessName string, sess session.SessionStore) int64 {
	sess_ := sess.Get(sessName)
	switch sess_.(type) {
	default:
		return 0
	case int64:
		return sess_.(int64)
	}
	return 0
}

func GetSessRestricted(sess session.SessionStore) int64 {
	sessRestricted := sess.Get("restricted")
	switch sessRestricted.(type) {
	default:
		return 0
	case int64:
		return sessRestricted.(int64)
	}
	return 0
}

func GetSessPublicKey(sess session.SessionStore) string {
	sessPublicKey := sess.Get("public_key")
	switch sessPublicKey.(type) {
	default:
		return ""
	case string:
		return sessPublicKey.(string)
	}
	return ""
}

// ключ в сессии хранится до того момента, пока юзер его не сменит
// т.е. ключ там лежит до тех пор, пока юзер играется и еще не начал пользоваться Dcoin-ом
func GetSessPrivateKey(w http.ResponseWriter, r *http.Request) string {
	sess, _ := globalSessions.SessionStart(w, r)
	defer sess.SessionRelease(w)
	sessPrivateKey := sess.Get("private_key")
	switch sessPrivateKey.(type) {
	default:
		return ""
	case string:
		return sessPrivateKey.(string)
	}
	return ""
}

func SetLang(w http.ResponseWriter, r *http.Request, lang int) {
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "lang", Value: strconv.Itoa(lang), Expires: expiration}
	http.SetCookie(w, &cookie)
}

// если в lang прислали какую-то гадость
func CheckLang(lang int) bool {
	for _, v := range consts.LangMap {
		if lang == v {
			return true
		}
	}
	return false
}

func GetLang(w http.ResponseWriter, r *http.Request, parameters map[string]string) int {
	var lang int = 1
	lang = utils.StrToInt(parameters["lang"])
	if !CheckLang(lang) {
		if langCookie, err := r.Cookie("lang"); err == nil {
			lang, _ = strconv.Atoi(langCookie.Value)
		}
	}
	if !CheckLang(lang) {
		al := r.Header.Get("Accept-Language") // en-US,en;q=0.5
		log.Debug("Accept-Language: %s", r.Header.Get("Accept-Language"))
		if len(al) >= 2 {
			if _, ok := consts.LangMap[al[:2]]; ok {
				lang = consts.LangMap[al[:2]]
			}
		}
	}
	return lang
}

func makeTemplate(html, name string, tData interface{}) (string, error) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("makeTemplate Recovered", r)
			fmt.Println(r)
		}
	}()

	data, err := static.Asset("static/" + html + ".html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	signatures, err := static.Asset("static/signatures.html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	alert_success, err := static.Asset("static/alert_success.html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	funcMap := template.FuncMap{
		"makeCurrencyName": func(currencyId int64) string {
			if currencyId >= 1000 {
				return ""
			} else {
				return "d"
			}
		},
		"div": func(a, b interface{}) float64 {
			return utils.InterfaceToFloat64(a) / utils.InterfaceToFloat64(b)
		},
		"mult": func(a, b interface{}) float64 {
			return utils.InterfaceToFloat64(a) * utils.InterfaceToFloat64(b)
		},
		"round": func(a interface{}, num int) float64 {
			return utils.Round(utils.InterfaceToFloat64(a), num)
		},
		"len": func(s []map[string]string) int {
			return len(s)
		},
		"lenMap": func(s map[string]string) int {
			return len(s)
		},
		"sum": func(a, b interface{}) float64 {
			return utils.InterfaceToFloat64(a) + utils.InterfaceToFloat64(b)
		},
		"minus": func(a, b interface{}) float64 {
			return utils.InterfaceToFloat64(a) - utils.InterfaceToFloat64(b)
		},
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
		"js": func(s string) template.JS {
			return template.JS(s)
		},
		"join": func(s []string, sep string) string {
			return strings.Join(s, sep)
		},
		"strToInt64": func(text string) int64 {
			return utils.StrToInt64(text)
		},
		"strToInt": func(text string) int {
			return utils.StrToInt(text)
		},
		"bin2hex": func(text string) string {
			return string(utils.BinToHex([]byte(text)))
		},
		"int64ToStr": func(text int64) string {
			return utils.Int64ToStr(text)
		},
		"intToStr": func(text int) string {
			return utils.IntToStr(text)
		},
		"intToInt64": func(text int) int64 {
			return int64(text)
		},
		"rand": func() int {
			return utils.RandInt(0, 99999999)
		},
		"append": func(args ...interface{}) string {
			var result string
			for _, value := range args {
				switch value.(type) {
				case int64:
					result += utils.Int64ToStr(value.(int64))
				case float64:
					result += utils.Float64ToStr(value.(float64))
				case string:
					result += value.(string)
				}
			}
			return result
		},
		"replaceCurrency": func(text, name string) string {
			return strings.Replace(text, "[currency]", name, -1)
		},
		"replaceCurrencyName": func(text, name string) string {
			return strings.Replace(text, "[currency]", "D"+name, -1)
		},
		"cfCategoryLang": func(lang map[string]string, name string) string {
			return lang["cf_category_"+name]
		},
		"progressBarLang": func(lang map[string]string, name string) string {
			return lang["progress_bar_pct_"+name]
		},
		"checkProjectPs": func(ProjectPs map[string]string, id string) bool {
			if len(ProjectPs["ps"+id]) > 0 {
				return true
			} else {
				return false
			}
		},
		"cfPageTypeLang": func(lang map[string]string, name string) string {
			return lang["cf_"+name]
		},
		"notificationsLang": func(lang map[string]string, name string) string {
			return lang["notifications_"+name]
		},
		"issuffix": func(text, name string) bool {
			return strings.HasSuffix(text,name)
		},

	}
	t := template.Must(template.New("template").Funcs(funcMap).Parse(string(data)))
	t = template.Must(t.Parse(string(alert_success)))
	t = template.Must(t.Parse(string(signatures)))
	b := new(bytes.Buffer)
	err = t.ExecuteTemplate(b, name, tData)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return b.String(), nil
}

func (c *Controller) GetParameters() (map[string]string, error) {
	parameters := make(map[string]string)
	if len(c.r.PostFormValue("parameters")) > 0 {
		parameters_ := make(map[string]interface{})
		err := json.Unmarshal([]byte(c.r.PostFormValue("parameters")), &parameters_)
		if err != nil {
			return parameters, utils.ErrInfo(err)
		}
		log.Debug("parameters_=", parameters_)
		for k, v := range parameters_ {
			parameters[k] = utils.InterfaceToStr(v)
		}
	}
	return parameters, nil
}

func userLock(userId int64) error {
	var affect int64
	var err error
	// даем время, чтобы lock освободился от другого запроса
	for i := 0; i < 4; i++ {
		affect, err = utils.DB.ExecSqlGetAffect(`UPDATE e_users SET lock = ? WHERE id = ? AND lock = 0`, utils.Time(), userId)
		if affect > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if affect == 0 {
		return errors.New("queue error")
	}
	return err
}

func NewForexOrder(userId int64, amount, sellRate float64, sellCurrencyId, buyCurrencyId int64, orderType string, eCommission float64) error {
	log.Debug("userId: %v / amount: %v / sellRate: %v / sellCurrencyId: %v / buyCurrencyId: %v / orderType: %v ", userId, amount, sellRate, sellCurrencyId, buyCurrencyId, orderType)
	curTime := utils.Time()
	err := userLock(userId)
	if err != nil {
		return utils.ErrInfo(err)
	}
	commission := utils.StrToFloat64(utils.ClearNull(utils.Float64ToStr(amount*(eCommission/100)), 6))
	newAmount := utils.StrToFloat64(utils.ClearNull(utils.Float64ToStr(amount-commission), 6))
	if newAmount < commission {
		commission = 0
		newAmount = amount
	}
	log.Debug("newAmount: %v / commission: %v ", newAmount, commission)
	if commission > 0 {
		userAmount := utils.EUserAmountAndProfit(1, sellCurrencyId)
		newAmount_ := userAmount + commission
		// наисляем комиссию системе
		err = utils.UpdEWallet(1, sellCurrencyId, utils.Time(), newAmount_, true)
		if err != nil {
			return utils.ErrInfo(err)
		}
		// и сразу вычитаем комиссию с кошелька юзера
		userAmount = utils.EUserAmountAndProfit(userId, sellCurrencyId)
		err = utils.DB.ExecSql("UPDATE e_wallets SET amount = ?, last_update = ? WHERE user_id = ? AND currency_id = ?", userAmount-commission, utils.Time(), userId, sellCurrencyId)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	// обратный курс. нужен для поиска по ордерам
	reverseRate := utils.StrToFloat64(utils.ClearNull(utils.Float64ToStr(1/sellRate), 6))
	log.Debug("sellRate %v", sellRate)
	log.Debug("reverseRate %v", reverseRate)

	var totalBuyAmount, totalSellAmount float64
	if orderType == "buy" {
		totalBuyAmount = newAmount * reverseRate
	} else {
		totalSellAmount = newAmount
	}

	var debit float64
	var prevUserId int64
	// берем из БД только те ордеры, которые удовлетворяют нашим требованиям
	rows, err := utils.DB.Query(utils.DB.FormatQuery(`
				SELECT id, user_id, amount, sell_rate, buy_currency_id, sell_currency_id
				FROM e_orders
				WHERE buy_currency_id = ? AND
							 sell_rate >= ? AND
							 sell_currency_id = ?  AND
							 del_time = 0 AND
							 empty_time = 0
				ORDER BY sell_rate DESC
				`), sellCurrencyId, reverseRate, buyCurrencyId)
	if err != nil {
		return utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var rowId, rowUserId, rowBuyCurrencyId, rowSellCurrencyId int64
		var rowAmount, rowSellRate float64
		err = rows.Scan(&rowId, &rowUserId, &rowAmount, &rowSellRate, &rowBuyCurrencyId, &rowSellCurrencyId)
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("rowId: %v / rowUserId: %v / rowAmount: %v / rowSellRate: %v / rowBuyCurrencyId: %v / rowSellCurrencyId: %v", rowId, rowUserId, rowAmount, rowSellRate, rowBuyCurrencyId, rowSellCurrencyId)

		// чтобы ордеры одного и тоже юзера не вызывали стопор, пропускаем все его оредера
		if rowUserId == prevUserId {
			continue
		}
		// блокируем юзера, чей ордер взяли, кроме самого себя
		if rowUserId != userId {
			lockErr := userLock(rowUserId)
			if lockErr != nil {
				log.Error("%v", utils.ErrInfo(err))
				prevUserId = rowUserId
				continue
			}
		}
		if orderType == "buy" {
			// удовлетворит ли данный ордер наш запрос целиком
			if rowAmount >= totalBuyAmount {
				debit = totalBuyAmount
				log.Debug("order ENDED")
			} else {
				debit = rowAmount
			}
		} else {
			// удовлетворит ли данный ордер наш запрос целиком
			if rowAmount/rowSellRate >= totalSellAmount {
				debit = totalSellAmount * rowSellRate
			} else {
				debit = rowAmount
			}
		}
		log.Debug("totalBuyAmount: %v / debit: %v", totalBuyAmount, debit)
		if rowAmount-debit < 0.01 { // ордер опустошили
			err = utils.DB.ExecSql("UPDATE e_orders SET amount = 0, empty_time = ? WHERE id = ?", curTime, rowId)
			if err != nil {
				return utils.ErrInfo(err)
			}
		} else {
			// вычитаем забранную сумму из ордера
			err = utils.DB.ExecSql("UPDATE e_orders SET amount = amount - ? WHERE id = ?", debit, rowId)
			if err != nil {
				return utils.ErrInfo(err)
			}
		}
		mySellRate := utils.ClearNull(utils.Float64ToStr(1/rowSellRate), 6)
		myAmount := debit * utils.StrToFloat64(mySellRate)
		eTradeSellCurrencyId := sellCurrencyId
		eTradeBuyCurrencyId := buyCurrencyId

		// для истории сделок
		err = utils.DB.ExecSql("INSERT INTO e_trade ( user_id, sell_currency_id, sell_rate, amount, buy_currency_id, time, main ) VALUES ( ?, ?, ?, ?, ?, ?, 1 )", userId, eTradeSellCurrencyId, mySellRate, myAmount, eTradeBuyCurrencyId, curTime)
		if err != nil {
			return utils.ErrInfo(err)
		}

		// тот, чей ордер обрабатываем
		err = utils.DB.ExecSql("INSERT INTO e_trade ( user_id, sell_currency_id, sell_rate, amount, buy_currency_id, time ) VALUES ( ?, ?, ?, ?, ?, ? )", rowUserId, eTradeBuyCurrencyId, rowSellRate, debit, eTradeSellCurrencyId, curTime)
		if err != nil {
			return utils.ErrInfo(err)
		}

		// ==== Продавец валюты (тот, чей ордер обработали) ====
		// сколько продавец данного ордера продал валюты
		sellerSellAmount := debit

		// сколько продавец получил buy_currency_id с продажи суммы $seller_sell_amount по его курсу
		sellerBuyAmount := sellerSellAmount * (1 / rowSellRate)

		// начисляем валюту, которую продавец получил (R)
		userAmount := utils.EUserAmountAndProfit(rowUserId, rowBuyCurrencyId)
		newAmount_ := userAmount + sellerBuyAmount
		err = utils.UpdEWallet(rowUserId, rowBuyCurrencyId, utils.Time(), newAmount_, true)
		if err != nil {
			return utils.ErrInfo(err)
		}

		// ====== Покупатель валюты (наш юзер) ======

		// списываем валюту, которую мы продали (R)
		userAmount = utils.EUserAmountAndProfit(userId, rowBuyCurrencyId)
		newAmount_ = userAmount - sellerBuyAmount
		err = utils.UpdEWallet(userId, rowBuyCurrencyId, utils.Time(), newAmount_, true)
		if err != nil {
			return utils.ErrInfo(err)
		}

		// начисляем валюту, которую мы получили (U)
		userAmount = utils.EUserAmountAndProfit(userId, rowSellCurrencyId)
		newAmount_ = userAmount + sellerSellAmount
		err = utils.UpdEWallet(userId, rowSellCurrencyId, utils.Time(), newAmount_, true)
		if err != nil {
			return utils.ErrInfo(err)
		}

		if orderType == "buy" {
			totalBuyAmount -= rowAmount
			if totalBuyAmount <= 0 {
				userUnlock(rowUserId)
				log.Debug("total_buy_amount == 0 break")
				break // проход по ордерам прекращаем, т.к. наш запрос удовлетворен
			}
		} else {
			totalSellAmount -= rowAmount / rowSellRate
			if totalSellAmount <= 0 {
				userUnlock(rowUserId)
				log.Debug("total_sell_amount == 0 break")
				break // проход по ордерам прекращаем, т.к. наш запрос удовлетворен
			}
		}
		if rowUserId != userId {
			userUnlock(rowUserId)
		}
	}

	log.Debug("totalBuyAmount: %v / orderType: %v / sellRate: %v / reverseRate: %v", totalBuyAmount, orderType, sellRate, reverseRate)

	// если после прохода по всем имеющимся ордерам мы не набрали нужную сумму, то создаем свой ордер
	if totalBuyAmount > 0 || totalSellAmount > 0 {
		var newOrderAmount float64
		if orderType == "buy" {
			newOrderAmount = totalBuyAmount * sellRate
		} else {
			newOrderAmount = totalSellAmount
		}
		log.Debug("newOrderAmount: %v", newOrderAmount)
		if newOrderAmount >= 0.000001 {
			err = utils.DB.ExecSql("INSERT INTO e_orders ( time, user_id, sell_currency_id, sell_rate, begin_amount, amount, buy_currency_id ) VALUES ( ?, ?, ?, ?, ?, ?, ? )", curTime, userId, sellCurrencyId, sellRate, newOrderAmount, newOrderAmount, buyCurrencyId)
			if err != nil {
				return utils.ErrInfo(err)
			}

			// вычитаем с баланса сумму созданного ордера
			userAmount := utils.EUserAmountAndProfit(userId, sellCurrencyId)
			err = utils.DB.ExecSql("UPDATE e_wallets SET amount = ?, last_update = ? WHERE user_id = ? AND currency_id = ?", userAmount-newOrderAmount, utils.Time(), userId, sellCurrencyId)
			if err != nil {
				return utils.ErrInfo(err)
			}
		}
	}

	userUnlock(userId)

	return nil
}

func userUnlock(userId int64) {
	utils.DB.ExecSql("UPDATE e_users SET lock = 0 WHERE id = ?", userId)
}

type CurrencyPct struct {
	Name       string
	Miner      float64
	User       float64
	MinerBlock float64
	UserBlock  float64
	MinerSec   float64
	UserSec    float64
}