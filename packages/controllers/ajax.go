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
	"fmt"
	"net/http"
	"regexp"

	"github.com/DayLightProject/go-daylight/packages/lib"
	"github.com/DayLightProject/go-daylight/packages/utils"
	qrcode "github.com/skip2/go-qrcode"
)

func Ajax(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("ajax Recovered", r)
			fmt.Println("ajax Recovered", r)
		}
	}()
	if qr := r.FormValue("qr"); len(qr) > 0 {
		if lib.IsValidAddress(qr) {
			png, _ := qrcode.Encode(qr, qrcode.Medium, 170)
			w.Header().Set("Content-Type", "image/png")
			w.Write(png)
		}
		return
	}
	log.Debug("Ajax")
	sess, err := globalSessions.SessionStart(w, r)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer sess.SessionRelease(w)
	sessWalletId := GetSessWalletId(sess)
	sessCitizenId := GetSessCitizenId(sess)
	sessAddress := GetSessString(sess, "address")
	sessStateId := GetSessInt64("state_id", sess)

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
	c.SessAddress = sessAddress
	c.SessStateId = sessStateId

	if dbInit {
		//c.DCDB, err = utils.NewDbConnect(configIni)

		c.DCDB = utils.DB

		if utils.DB == nil || utils.DB.DB == nil {
			log.Error("utils.DB == nil")
			dbInit = false
		}
	}
	if sessStateId > 0 {
		stateName, err := c.GetStateName(sessStateId)
		if err != nil {
			log.Error("%v", err)
		}
		c.StateName = stateName
		c.StateId = sessStateId
		c.StateIdStr = utils.Int64ToStr(sessStateId)
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
		config, err := c.GetNodeConfig()
		if err != nil {
			log.Error("%v", err)
		}
		c.NodeConfig = config

	}

	r.ParseForm()
	if jsonName := r.FormValue(`json`); len(jsonName) > 0 && isPage(jsonName, TJson) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(CallJson(c, jsonName))
		return
	}

	w.Header().Set("Content-type", "text/html")

	controllerName := r.FormValue("controllerName")
	log.Debug("controllerName=", controllerName)

	html := ""

	//w.Header().Set("Access-Control-Allow-Origin", "*")
	// Общие контролы для двух проверок
	pages := "SignIn|UpdateDcoin|AlertFromAdmin|FreecoinProcess|RestartDb|ReloadDb|DebugInfo|CheckSetupPassword|AcceptNewKeyStatus|availableKeys|CfCatalog|CfPagePreview|CfStart|CheckNode|GetBlock|GetMinerData|GetMinerDataMap|GetSellerData|Index|IndexCf|InstallStep0|InstallStep1|InstallStep2|Login|SynchronizationBlockchain|UpdatingBlockchain|Menu|SignUpInPool|SignLogin"
	// Почему CfCatalog,CfPagePreview,CfStart,Index,IndexCf,InstallStep0,InstallStep1,
	// InstallStep2,Login,UpdatingBlockchain были только во втором случае? Похоже не нужны больше.

	if ok, _ := regexp.MatchString(`^(?i)`+pages+`|GetServerTime|TxStatus|AnonymHistory|RewritePrimaryKeySave|SendPromisedAmountToPool|SaveEmailAndSendTestMess|sendMobile|rewritePrimaryKey|EImportData|EDataBaseDump|Update|exchangeAdmin|exchangeSupport|exchangeUser|ETicket|newPhoto|NodeConfigControl|SaveDecryptComment|EncryptChatMessage|GetChatMessages|SendToTheChat|SaveToken|SendToPool|ClearDbLite|ClearDb|UploadVideo|DcoinKey|PoolAddUsers|SaveQueue|AlertMessage|SaveHost|PoolDataBaseDump|GenerateNewPrimaryKey|GenerateNewNodeKey|SaveNotifications|ProgressBar|MinersMap|EncryptComment|Logout|SaveVideo|SaveShopData|SaveRaceCountry|MyNoticeData|HolidaysList|ClearVideo|CheckCfCurrency|WalletsListCfProject|SendTestEmail|SendSms|SaveUserCoords|SaveGeolocation|SaveEmailSms|Profile|DeleteVideo|CropPhoto$`, controllerName); !ok {
		html = "Access denied 0"
	} else {
		if utils.Mobile() { // На IOS можно сгенерить ключ без сессии
			pages += "|DcoinKey"
		}
		if ok, _ := regexp.MatchString(`^(?i)`+pages+`$`, controllerName); !ok && c.SessWalletId <= 0 && c.SessCitizenId <= 0 && len(c.SessAddress) == 0 {
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
