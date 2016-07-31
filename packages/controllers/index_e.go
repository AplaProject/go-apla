package controllers

import (
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

type indexE struct {
	MyWallets         []map[string]string
	Lang              map[string]string
	Nav               template.JS
	UserId            int64
	EHost             string
	AnalyticsDisabled string
}

func EStaticFile(w http.ResponseWriter, r *http.Request) {
	static_file, err := utils.DB.Single(`SELECT value FROM e_config WHERE name='static_file'`).Bytes()
	if err != nil {
		log.Error("%v", err)
	}
	w.Write(static_file)
}

func IndexE(w http.ResponseWriter, r *http.Request) {

	var err error

	if utils.DB != nil && utils.DB.DB != nil {

		sess, _ := globalSessions.SessionStart(w, r)
		defer sess.SessionRelease(w)

		c := new(Controller)
		c.r = r
		c.SessUserId = GetSessEUserId(sess)
		c.DCDB = utils.DB

		r.ParseForm()

		c.Parameters, err = c.GetParameters()
		log.Debug("parameters=", c.Parameters)

		lang := GetLang(w, r, c.Parameters)
		log.Debug("lang", lang)
		c.Lang = globalLangReadOnly[lang]

		var myWallets []map[string]string
		if c.SessUserId > 0 {
			myWallets, err = c.getMyWallets()
			if err != nil {
				w.Write([]byte(utils.ErrInfo(err).Error()))
				log.Error("%v", err)
				return
			}
		}

		c.EConfig, err = c.GetMap(`SELECT * FROM e_config`, "name", "value")
		if err != nil {
			log.Error("%v", err)
		}
		eHost := c.EConfig["domain"]
		if len(eHost) == 0 {
			http_host, err := c.Single(`SELECT http_host FROM config`).String()
			if err != nil {
				w.Write([]byte(utils.ErrInfo(err).Error()))
				log.Error("%v", err)
				return
			}
			re := regexp.MustCompile(`^https?:\/\/([0-9a-z\_\.\-:]+)\/?`)
			match := re.FindStringSubmatch(http_host)
			catalog := strings.Replace(c.EConfig["catalog"], "/", "", -1)
			if len(match) != 0 {
				c.EConfig["catalog"] = catalog
				eHost = match[1] + "/" + catalog + "/"
			} else if len(http_host) == 0 {
				eHost = r.Host + "/" + catalog + "/"
			}
		}
		analyticsDisabled, err := utils.DB.Single(`SELECT analytics_disabled FROM config`).String()
		if err != nil {
			log.Error("%v", err)
		}

		data, err := static.Asset("static/templates/index_e.html")
		if err != nil {
			w.Write([]byte(utils.ErrInfo(err).Error()))
			log.Error("%v", err)
			return
		}
		t := template.New("template")
		t, err = t.Parse(string(data))
		if err != nil {
			w.Write([]byte(utils.ErrInfo(err).Error()))
			log.Error("%v", err)
			return
		}
		b := new(bytes.Buffer)
		err = t.Execute(b, &indexE{MyWallets: myWallets, Lang: c.Lang, UserId: c.SessUserId, EHost: eHost, AnalyticsDisabled: analyticsDisabled})
		if err != nil {
			log.Error("%v", err)
			w.Write([]byte(utils.ErrInfo(err).Error()))
		} else {
			w.Write(b.Bytes())
		}

	}
}

func (c *Controller) getMyWallets() ([]map[string]string, error) {
	var myWallets []map[string]string
	eCurrency, err := c.GetAll(`SELECT name as currency_name, id FROM e_currency ORDER BY sort_id ASC`, -1)
	if err != nil {
		return myWallets, utils.ErrInfo(err)
	}
	for _, data := range eCurrency {
		wallet, err := c.OneRow("SELECT * FROM e_wallets WHERE user_id  =  ? AND currency_id  =  ?", c.SessUserId, data["id"]).String()
		if err != nil {
			return myWallets, utils.ErrInfo(err)
		}
		if len(wallet) > 0 {
			amount := utils.StrToFloat64(wallet["amount"])
			profit, err := utils.DB.CalcProfitGen(utils.StrToInt64(wallet["currency_id"]), amount, 0, utils.StrToInt64(wallet["last_update"]), utils.Time(), "wallet")
			if err != nil {
				return myWallets, utils.ErrInfo(err)
			}
			myWallets = append(myWallets, map[string]string{"amount": utils.ClearNull(utils.Float64ToStr(amount+profit), 2), "currency_name": data["currency_name"], "last_update": wallet["last_update"]})
		}
	}

	return myWallets, nil
}
