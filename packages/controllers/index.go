package controllers

import (
	"bytes"
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

type index struct {
	DbOk              bool
	Lang              map[string]string
	Key               string
	SetLang           string
	IOS               bool
	Android           bool
	Mobile            bool
	ShowIOSMenu       bool
}

func Index(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	parameters_ := make(map[string]interface{})
	if len(r.PostFormValue("parameters")) > 0 {
		err := json.Unmarshal([]byte(r.PostFormValue("parameters")), &parameters_)
		if err != nil {
			log.Error("%v", err)
		}
		log.Debug("parameters_=%", parameters_)
	}
	parameters := make(map[string]string)
	for k, v := range parameters_ {
		parameters[k] = utils.InterfaceToStr(v)
	}

	lang := GetLang(w, r, parameters)

	sess, err := globalSessions.SessionStart(w, r)
	if err != nil {
		log.Error("%v", err)
	}
	defer sess.SessionRelease(w)

	sessCitizenId := GetSessCitizenId(sess)
	sessWalletId := GetSessWalletId(sess)

	var key string

	showIOSMenu := true
	// Когда меню не выдаем
	if utils.DB == nil || utils.DB.DB == nil {
		showIOSMenu = false
	}

	if sessCitizenId == 0 && sessWalletId == 0 {
		showIOSMenu = false
	}

	if showIOSMenu && utils.DB != nil && utils.DB.DB != nil {
		blockData, err := utils.DB.GetInfoBlock()
		if err != nil {
			log.Error("%v", err)
		}
		wTime := int64(12)
		wTimeReady := int64(2)
		log.Debug("wTime: %v / utils.Time(): %v / blockData[time]: %v", wTime, utils.Time(), utils.StrToInt64(blockData["time"]))
		// если время менее 12 часов от текущего, то выдаем не подвержденные, а просто те, что есть в блокчейне
		if utils.Time()-utils.StrToInt64(blockData["time"]) < 3600*wTime {
			lastBlockData, err := utils.DB.GetLastBlockData()
			if err != nil {
				log.Error("%v", err)
			}
			log.Debug("lastBlockData[lastBlockTime]: %v", lastBlockData["lastBlockTime"])
			log.Debug("time.Now().Unix(): %v", utils.Time())
			if utils.Time()-lastBlockData["lastBlockTime"] >= 3600*wTimeReady {
				showIOSMenu = false
			}
		} else {
			showIOSMenu = false
		}
	}
	if showIOSMenu && !utils.Mobile() {
		showIOSMenu = false
	}

	mobile := utils.Mobile()
	if ok, _ := regexp.MatchString("(?i)(iPod|iPhone|iPad|Android)", r.UserAgent()); ok {
		mobile = true
	}

	ios := utils.IOS()
	if ok, _ := regexp.MatchString("(?i)(iPod|iPhone|iPad)", r.UserAgent()); ok {
		ios = true
	}

	android := utils.Android()
	if ok, _ := regexp.MatchString("(?i)(Android)", r.UserAgent()); ok {
		android = true
	}

	key = strings.Replace(key, "\r", "\n", -1)
	key = strings.Replace(key, "\n\n", "\n", -1)
	key = strings.Replace(key, "\n", "\\\n", -1)

	setLang := r.FormValue("lang")


	funcMap := template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	data, err := static.Asset("static/index.html")
	t := template.New("template").Funcs(funcMap)
	t, err = t.Parse(string(data))
	if err != nil {
		log.Error("%v", err)
	}

	b := new(bytes.Buffer)
	err = t.Execute(b, &index{
		DbOk:        true,
		Lang:        globalLangReadOnly[lang],
		Key:         key,
		SetLang:     setLang,
		ShowIOSMenu: showIOSMenu,
		IOS:               ios,
		Android:           android,
		Mobile:            mobile})
	if err != nil {
		log.Error("%v", err)
	}
	w.Write(b.Bytes())
}
