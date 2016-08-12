package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"net/http"
	"regexp"
	"fmt"
)

func Ajax(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("ajax Recovered", r)
			fmt.Println("ajax Recovered", r)
		}
	}()
	log.Debug("Ajax")
	w.Header().Set("Content-type", "text/html")

	sess, err := globalSessions.SessionStart(w, r)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer sess.SessionRelease(w)
	sessWalletId := GetSessWalletId(sess)
	sessCitizenId := GetSessCitizenId(sess)
	log.Debug("sessWalletId", sessWalletId)
	log.Debug("sessCitizenId", sessCitizenId)

	c := new(Controller)
	c.r = r
	c.w = w
	c.sess = sess
	dbInit := false
	if len(configIni["db_user"]) > 0 || configIni["db_type"] == "sqlite" {
		dbInit = true
	}

	c.SessWalletId = sessWalletId
	c.SessCitizenId = sessCitizenId
	if dbInit {
		//c.DCDB, err = utils.NewDbConnect(configIni)

		c.DCDB = utils.DB

		if utils.DB == nil || utils.DB.DB == nil {
			log.Error("utils.DB == nil")
			dbInit = false
		}
	}
	c.dbInit = dbInit

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

	if dbInit {
		myNotice, err := c.DCDB.GetMyNoticeData(sessCitizenId, sessWalletId, globalLangReadOnly[lang])
		if err != nil {
			log.Error("%v", err)
		}
		c.MyNotice = myNotice
		config, err := c.GetNodeConfig()
		if err != nil {
			log.Error("%v", err)
		}
		c.NodeConfig = config
		// валюты
		c.CurrencyList, err = c.GetCurrencyList(false)
		if err != nil {
			log.Error("%v", err)
		}
	}

	r.ParseForm()
	controllerName := r.FormValue("controllerName")
	log.Debug("controllerName=", controllerName)

	html := ""

	//w.Header().Set("Access-Control-Allow-Origin", "*")
	// Общие контролы для двух проверок
	pages := "SignIn|UpdateDcoin|AlertFromAdmin|FreecoinProcess|RestartDb|ReloadDb|DebugInfo|CheckSetupPassword|AcceptNewKeyStatus|availableKeys|CfCatalog|CfPagePreview|CfStart|CheckNode|GetBlock|GetMinerData|GetMinerDataMap|GetSellerData|Index|IndexCf|InstallStep0|InstallStep1|InstallStep2|Login|SynchronizationBlockchain|UpdatingBlockchain|Menu|SignUpInPool|SignLogin"
	// Почему CfCatalog,CfPagePreview,CfStart,Index,IndexCf,InstallStep0,InstallStep1,
	// InstallStep2,Login,UpdatingBlockchain были только во втором случае? Похоже не нужны больше.
	
	if ok, _ := regexp.MatchString(`^(?i)` + pages + `|GetServerTime|TxStatus|AnonymHistory|RewritePrimaryKeySave|SendPromisedAmountToPool|SaveEmailAndSendTestMess|sendMobile|rewritePrimaryKey|EImportData|EDataBaseDump|Update|exchangeAdmin|exchangeSupport|exchangeUser|ETicket|newPhoto|NodeConfigControl|SaveDecryptComment|EncryptChatMessage|GetChatMessages|SendToTheChat|SaveToken|SendToPool|ClearDbLite|ClearDb|UploadVideo|DcoinKey|PoolAddUsers|SaveQueue|AlertMessage|SaveHost|PoolDataBaseDump|GenerateNewPrimaryKey|GenerateNewNodeKey|SaveNotifications|ProgressBar|MinersMap|EncryptComment|Logout|SaveVideo|SaveShopData|SaveRaceCountry|MyNoticeData|HolidaysList|ClearVideo|CheckCfCurrency|WalletsListCfProject|SendTestEmail|SendSms|SaveUserCoords|SaveGeolocation|SaveEmailSms|Profile|DeleteVideo|CropPhoto$`, controllerName); !ok {
		html = "Access denied 0"
	} else {
		if utils.Mobile() { // На IOS можно сгенерить ключ без сессии
			pages += "|DcoinKey"
		}
		if ok, _ := regexp.MatchString(`^(?i)`+pages+`$`, controllerName); !ok && c.SessWalletId <= 0 && c.SessCitizenId <= 0 {
			html = "Access denied 1"
		} else {
			// без БД будет выдавать панику
			if ok, _ := regexp.MatchString(`^(?i)GetChatMessages$`, controllerName); ok && !dbInit {
				html = "Please wait. nill dbInit"
			} else {
				// вызываем контроллер в зависимости от шаблона
				html, err = CallController(c, controllerName)
				if err != nil {
					log.Error("ajax error: %v", err)
				}
			}
		}
	}
	w.Write([]byte(html))

}
