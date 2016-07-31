package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html/template"
	"net/http"
	"regexp"
)

type contentE struct {
	CfUrl  string
	Lang   string
	Nav    template.JS
	CfLang map[string]string
}

func ContentE(w http.ResponseWriter, r *http.Request) {

	var err error
	if utils.DB != nil && utils.DB.DB != nil {

		sess, _ := globalSessions.SessionStart(w, r)
		defer sess.SessionRelease(w)

		c := new(Controller)
		c.r = r
		c.SessUserId = GetSessEUserId(sess)
		c.DCDB = utils.DB

		r.ParseForm()
		tplName := r.FormValue("page")

		c.Parameters, err = c.GetParameters()
		log.Debug("parameters=", c.Parameters)

		lang := GetLang(w, r, c.Parameters)
		c.Lang = globalLangReadOnly[lang]
		c.LangInt = int64(lang)
		if lang == 42 {
			c.TimeFormat = "2006-01-02 15:04:05"
		} else {
			c.TimeFormat = "2006-02-01 15:04:05"
		}
		// если в параметрах пришел язык, то установим его
		newLang := utils.StrToInt(c.Parameters["lang"])
		if newLang > 0 {
			log.Debug("newLang", newLang)
			SetLang(w, r, newLang)
		}

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
		html := ""
		if ok, _ := regexp.MatchString(`^(?i)EPages|emain|EMyOrders|EMyHistory|EMyFinance|EMySupport`, tplName); !ok {
			html = "Access denied"
		} else {
			// вызываем контроллер в зависимости от шаблона
			html, err = CallController(c, tplName)
			if err != nil {
				log.Error("%v", err)
			}
		}
		w.Write([]byte(html))
	}
}
