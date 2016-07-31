package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html/template"
	"net/http"
	"regexp"
)

type contentCf struct {
	CfUrl  string
	Lang   string
	Nav    template.JS
	CfLang map[string]string
}

func ContentCf(w http.ResponseWriter, r *http.Request) {

	c := new(Controller)
	c.r = r
	dbInit := false
	if len(configIni["db_user"]) > 0 || (configIni["db_type"] == "sqlite") {
		dbInit = true
	}
	if dbInit {
		var err error
		c.DCDB = utils.DB
		if c.DCDB.DB == nil {
			log.Error("utils.DB == nil")
			dbInit = false
		}
		// отсутвие таблы выдаст ошибку, значит процесс инсталяции еще не пройден и надо выдать 0-й шаг
		_, err = c.DCDB.Single("SELECT progress FROM install").String()
		if err != nil {
			log.Error("%v", err)
			dbInit = false
		}

		cfUrl, err := c.GetCfUrl()
		if len(cfUrl) == 0 {
			w.Write([]byte("die"))
			return
		}
		//c.CfUrl = cfUrl

		r.ParseForm()
		tplName := r.FormValue("tpl_name")

		c.Parameters, err = c.GetParameters()
		log.Debug("parameters=", c.Parameters)

		lang := GetLang(w, r, c.Parameters)
		c.Lang = globalLangReadOnly[lang]
		log.Debug("c.Lang:", c.Lang)
		c.LangInt = int64(lang)

		// если в параметрах пришел язык, то установим его
		newLang := utils.StrToInt(c.Parameters["lang"])
		if newLang > 0 {
			log.Debug("newLang", newLang)
			SetLang(w, r, newLang)
		}

		config, err := c.GetNodeConfig()
		sess, _ := globalSessions.SessionStart(w, r)
		defer sess.SessionRelease(w)
		c.SessUserId = GetSessUserId(sess)
		if config["pool_admin_user_id"] != "0" && config["pool_admin_user_id"] != utils.Int64ToStr(c.SessUserId) && config["pool_tech_works"] != "1" {
			tplName = "pool_tech_works"
		} else if len(tplName) > 0 {
			if ok, _ := regexp.MatchString("^[\\w]{1,30}$", tplName); !ok {
				tplName = "cfCatalog"
			}
		} else {
			tplName = "cfCatalog"
		}

		html := ""
		if ok, _ := regexp.MatchString(`^(?i)CfPagePreview|CfCatalog|AddCfProjectData|CfProjectChangeCategory|NewCfProject|MyCfProjects|DelCfProject|DelCfFunding|CfStart$`, tplName); !ok {
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
