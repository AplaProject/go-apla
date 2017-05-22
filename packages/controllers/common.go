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
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/session"
	"github.com/op/go-logging"
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
	ContentInc       bool
	Periods          map[int64]string
	Alert            string
	SessStateId      int64
	StateName        string
	StateId          int64
	StateIdStr       string
	SessCitizenId    int64
	SessWalletId     int64
	SessAddress      string
	MyNotice         map[string]string
	Parameters       map[string]string
	TimeFormat       string
	NodeAdmin        bool
	NodeConfig       map[string]string
	CurrencyList     map[int64]string
	ConfirmedBlockId int64
	Data             *CommonPage
}

var (
	configIni      map[string]string
	globalSessions *session.Manager
	// в гоурутинах используется только для чтения
	// In gourutin is used only for reading
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
	// We monitor config.ini for changes
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
	// the name of exported method must begin with a capital letter
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

func GetSessString(sess session.SessionStore, name string) string {
	sessVal := sess.Get(name)
	switch sessVal.(type) {
	case string:
		return sessVal.(string)
	}
	return ""
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

func SetLang(w http.ResponseWriter, r *http.Request, lang int) {
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "lang", Value: strconv.Itoa(lang), Expires: expiration}
	http.SetCookie(w, &cookie)
}

// если в lang прислали какую-то гадость
// If some muck was sent in the lang
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
		"replaceBr": func(text string) template.HTML {
			text = strings.Replace(text, `\n`, "<br>", -1)
			text = strings.Replace(text, `\t`, "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
			return template.HTML(text)
		},
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
			return strings.HasSuffix(text, name)
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
