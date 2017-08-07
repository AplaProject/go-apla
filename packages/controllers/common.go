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
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/EGaaS/go-egaas-mvp/packages/api"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	bconf "github.com/astaxie/beego/config"
	"github.com/astaxie/beego/session"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("controllers")

// Controller is the main controller's structure
type Controller struct {
	dbInit bool
	*sql.DCDB
	r                *http.Request
	w                http.ResponseWriter
	sess             session.SessionStore
	Lang             map[string]string
	TplName          string
	LangInt          int64
	ContentInc       bool
	Periods          map[int64]string
	Alert            string
	SessStateID      int64
	StateName        string
	StateID          int64
	StateIDStr       string
	SessCitizenID    int64
	SessWalletID     int64
	SessAddress      string
	MyNotice         map[string]string
	Parameters       map[string]string
	TimeFormat       string
	NodeAdmin        bool
	NodeConfig       map[string]string
	CurrencyList     map[int64]string
	ConfirmedBlockID int64
	Data             *CommonPage
}

var (
	globalSessions *session.Manager
	// In gourutin is used only for reading
	globalLangReadOnly map[int]map[string]string
)

// SessInit initializes sessions
func SessInit() {
	var err error

	globalSessions, err = session.NewManager("memory", `{"cookieName":"gosessionid","gclifetime":864000}`)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
	}
	api.SetSession(globalSessions)
	go globalSessions.GC()
}

func init() {
	flag.Parse()
	globalLangReadOnly = make(map[int]map[string]string)
	for _, v := range consts.LangMap {
		data, err := static.Asset(fmt.Sprintf("static/lang/%d.ini", v))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
		iniConf, err := bconf.NewConfigData("ini", []byte(data))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
		iniconf, err := iniConf.GetSection("default")
		globalLangReadOnly[v] = make(map[string]string)
		globalLangReadOnly[v] = iniconf
	}
}

// CallController calls the method with this name
func CallController(c *Controller, name string) (string, error) {
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

// CallMethod calls the method
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
		ierr, found := x[1].Interface().(error)
		var err error
		if found {
			err = ierr
		} else {
			err = nil
		}
		return x[0].Interface().(string), err
	}

	// return or panic, method not found of either type
	return "", fmt.Errorf("method not found")
}

// GetSessWalletID returns session's wallet id
func GetSessWalletID(sess session.SessionStore) int64 {
	sessUserID := sess.Get("wallet_id")
	log.Debug("sessUserId: %v", sessUserID)
	switch sessUserID.(type) {
	case int64:
		return sessUserID.(int64)
	case int:
		return int64(sessUserID.(int))
	case string:
		return converter.StrToInt64(sessUserID.(string))
	}
	return 0
}

// GetSessCitizenID returns session's citizen id
func GetSessCitizenID(sess session.SessionStore) int64 {
	sessUserID := sess.Get("citizen_id")
	log.Debug("sessUserId: %v", sessUserID)
	switch sessUserID.(type) {
	case int64:
		return sessUserID.(int64)
	case int:
		return int64(sessUserID.(int))
	case string:
		return converter.StrToInt64(sessUserID.(string))
	}
	return 0
}

// GetSessInt64 returns the integer value of the session key
func GetSessInt64(sessName string, sess session.SessionStore) int64 {
	val := sess.Get(sessName)
	switch val.(type) {
	case int64:
		return val.(int64)
	}
	return 0
}

// GetSessString returns the string value of the session key
func GetSessString(sess session.SessionStore, name string) string {
	sessVal := sess.Get(name)
	switch sessVal.(type) {
	case string:
		return sessVal.(string)
	}
	return ""
}

// GetSessPublicKey returns the session public key
func GetSessPublicKey(sess session.SessionStore) string {
	sessPublicKey := sess.Get("public_key")
	switch sessPublicKey.(type) {
	case string:
		return sessPublicKey.(string)
	}
	return ""
}

// SetLang sets lang cookie
func SetLang(w http.ResponseWriter, r *http.Request, lang int) {
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "lang", Value: strconv.Itoa(lang), Expires: expiration}
	http.SetCookie(w, &cookie)
}

// CheckLang checks if there is a language with such id
// If some muck was sent in the lang
func CheckLang(lang int) bool {
	for _, v := range consts.LangMap {
		if lang == v {
			return true
		}
	}
	return false
}

// GetLang returns the user's language
func GetLang(w http.ResponseWriter, r *http.Request, parameters map[string]string) int {
	lang := converter.StrToInt(parameters["lang"])
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
	alertSuccess, err := static.Asset("static/alert_success.html")
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
			}
			return "d"
		},
		"div": func(a, b interface{}) float64 {
			return converter.InterfaceToFloat64(a) / converter.InterfaceToFloat64(b)
		},
		"mult": func(a, b interface{}) float64 {
			return converter.InterfaceToFloat64(a) * converter.InterfaceToFloat64(b)
		},
		"round": func(a interface{}, num int) float64 {
			return converter.RoundWithPrecision(converter.InterfaceToFloat64(a), num)
		},
		"len": func(s []map[string]string) int {
			return len(s)
		},
		"lenMap": func(s map[string]string) int {
			return len(s)
		},
		"sum": func(a, b interface{}) float64 {
			return converter.InterfaceToFloat64(a) + converter.InterfaceToFloat64(b)
		},
		"minus": func(a, b interface{}) float64 {
			return converter.InterfaceToFloat64(a) - converter.InterfaceToFloat64(b)
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
			return converter.StrToInt64(text)
		},
		"strToInt": func(text string) int {
			return converter.StrToInt(text)
		},
		"bin2hex": func(text string) string {
			return string(converter.BinToHex([]byte(text)))
		},
		"int64ToStr": func(text int64) string {
			return converter.Int64ToStr(text)
		},
		"intToStr": func(text int) string {
			return converter.IntToStr(text)
		},
		"intToInt64": func(text int) int64 {
			return int64(text)
		},
		"rand": func() int {
			return crypto.RandInt(0, 99999999)
		},
		"append": func(args ...interface{}) string {
			var result string
			for _, value := range args {
				switch value.(type) {
				case int64:
					result += converter.Int64ToStr(value.(int64))
				case float64:
					result += converter.Float64ToStr(value.(float64))
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
			return len(ProjectPs["ps"+id]) > 0
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
	t = template.Must(t.Parse(string(alertSuccess)))
	t = template.Must(t.Parse(string(signatures)))
	b := new(bytes.Buffer)
	err = t.ExecuteTemplate(b, name, tData)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return b.String(), nil
}

// GetParameters returns the map of parameters
func (c *Controller) GetParameters() (map[string]string, error) {
	parameters := make(map[string]string)
	if len(c.r.PostFormValue("parameters")) > 0 {
		params := make(map[string]interface{})
		err := json.Unmarshal([]byte(c.r.PostFormValue("parameters")), &params)
		if err != nil {
			return parameters, utils.ErrInfo(err)
		}
		log.Debug("parameters_=", params)
		for k, v := range params {
			parameters[k] = converter.InterfaceToStr(v)
		}
	}
	return parameters, nil
}
