package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"net/http"
	"regexp"
)

func AjaxE(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("ajax Recovered", r)
			panic(r)
		}
	}()
	log.Debug("AjaxE")
	w.Header().Set("Content-type", "text/html")

	sess, err := globalSessions.SessionStart(w, r)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer sess.SessionRelease(w)
	sessUserId := GetSessEUserId(sess)
	log.Debug("sessUserId", sessUserId)

	c := new(Controller)
	c.r = r
	c.w = w
	c.sess = sess
	c.SessUserId = sessUserId

	if utils.DB == nil || utils.DB.DB == nil {
		log.Error("utils.DB == nil")
		w.Write([]byte("DB == nil"))
		return
	}
	c.DCDB = utils.DB

	c.Parameters, err = c.GetParameters()
	log.Debug("parameters=", c.Parameters)

	lang := GetLang(w, r, c.Parameters)
	log.Debug("lang", lang)
	c.Lang = globalLangReadOnly[lang]
	c.LangInt = int64(lang)
	if lang == 42 {
		c.TimeFormat = "2006-01-02 15:04:05"
	} else {
		c.TimeFormat = "2006-02-01 15:04:05"
	}

	r.ParseForm()
	controllerName := r.FormValue("controllerName")
	log.Debug("controllerName=", controllerName)

	html := ""

	c.EConfig, err = c.GetMap(`SELECT * FROM e_config`, "name", "value")
	if err != nil {
		log.Error("%v", err)
	}
	c.EURL = c.EConfig["domain"]
	if len(c.EURL) == 0 {
		eHost, err := c.Single(`SELECT http_host FROM config`).String()
		if err != nil {
			log.Error("%v", err)
		}
		eHost += c.EConfig["catalog"]
		c.EURL = eHost
	} else {
		c.EURL = "http://" + c.EURL + "/"
	}
	c.ECommission = utils.StrToMoney(c.EConfig["commission"])
	// валюты
	c.CurrencyList, err = c.GetCurrencyList(false)
	if err != nil {
		log.Error("%v", err)
	}
	if ok, _ := regexp.MatchString(`^(?i)EMenu|ETicket|EGateCP|EPayeerSign|EGatePayeer|EDelOrder|EWithdraw|EGetBalance|ESaveOrder|ESignUp|ELogin|ELogout|ESignLogin|ECheckSign|ERedirect|EInfo|EData|EGatePm|EGateIk$`, controllerName); !ok {
		html = "Access denied 0"
	} else {
		if ok, _ := regexp.MatchString(`^(?i)EMenu|ETicket|EPayeerSign|ESaveOrder|ESignUp|ELogin|ESignLogin|ECheckSign|ERedirect|EInfo|EData|EGatePm|EGateCP|EGateIk|EGatePayeer$`, controllerName); !ok && c.SessUserId <= 0 {
			html = "Access denied 1"
		} else {
			// вызываем контроллер в зависимости от шаблона
			log.Debug("controllerName %s", controllerName)
			html, err = CallController(c, controllerName)
			log.Debug("html %s", html)
			if err != nil {
				log.Error("ajax error: %v", err)
			}
		}
	}
	w.Write([]byte(html))

}
