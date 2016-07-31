package controllers

import (
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html/template"
	"os"
)

type progressBarPage struct {
	ProgressPct    int64
	Lang           map[string]string
	ProgressBar    map[string]int64
	ProgressBarPct map[string]int64
}

func (c *Controller) ProgressBar() (string, error) {

	if !c.dbInit {
		return "", nil
	}

	progressBarPct := make(map[string]int64)
	progressBarPct["begin"] = 10
	progressBarPct["change_key"] = 10
	progressBarPct["my_table"] = 5
	progressBarPct["upgrade_country"] = 3
	progressBarPct["upgrade_face_hash"] = 3
	progressBarPct["upgrade_profile_hash"] = 3
	progressBarPct["upgrade_face_coords"] = 3
	progressBarPct["upgrade_profile_coords"] = 3
	progressBarPct["upgrade_video"] = 3
	progressBarPct["upgrade_host"] = 3
	progressBarPct["upgrade_geolocation"] = 3
	progressBarPct["promised_amount"] = 5
	progressBarPct["commission"] = 3
	progressBarPct["tasks"] = 8
	progressBarPct["vote"] = 5
	progressBarPct["referral"] = 1

	progressBar := make(map[string]int64)

	// сменил ли юзер ключ
	changeKey, err := c.Single("SELECT log_id FROM users WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangePrimaryKey"}), 1, c.TimeFormat)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if (len(last_tx) > 0 && (len(last_tx[0]["queue_tx"]) > 0 || len(last_tx[0]["tx"]) > 0)) || changeKey > 0 {
		progressBar["change_key"] = 1
	}

	// есть ли в БД личная юзерсая таблица
	if c.Community {
		tables, err := c.GetAllTables()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if utils.InSliceString(utils.Int64ToStr(c.SessUserId)+"_my_table", tables) {
			progressBar["my_table"] = 1
		}
	} else {
		progressBar["my_table"] = 1
	}

	// апгрейд аккаунта
	myMinersId, err := c.GetMinerId(c.SessUserId)
	if myMinersId > 0 {
		progressBar["upgrade_country"] = 1
		progressBar["upgrade_face_hash"] = 1
		progressBar["upgrade_profile_hash"] = 1
		progressBar["upgrade_face_coords"] = 1
		progressBar["upgrade_profile_coords"] = 1
		progressBar["upgrade_video"] = 1
		progressBar["upgrade_host"] = 1
		progressBar["upgrade_geolocation"] = 1
	} else if c.SessRestricted == 0 {
		upgradeData, err := c.OneRow("SELECT user_id, race, country, geolocation, http_host as host, face_coords, profile_coords, video_url_id, video_type FROM " + c.MyPrefix + "my_table").String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(upgradeData["race"]) > 0 && len(upgradeData["country"]) > 0 {
			progressBar["upgrade_country"] = 1
		}
		if len(upgradeData["face_hash"]) > 0 {
			progressBar["upgrade_face_hash"] = 1
		}
		if len(upgradeData["profile_hash"]) > 0 {
			progressBar["upgrade_profile_hash"] = 1
		}
		if len(upgradeData["face_coords"]) > 0 {
			progressBar["upgrade_face_coords"] = 1
		}
		if len(upgradeData["profile_coords"]) > 0 {
			progressBar["upgrade_profile_coords"] = 1
		}
		if _, err := os.Stat(*utils.Dir + "public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.mp4"); os.IsExist(err) {
			if len(upgradeData["video_url_id"]) > 0 {
				progressBar["upgrade_video"] = 1
			}
		}
		if len(upgradeData["host"]) > 0 {
			progressBar["upgrade_host"] = 1
		}
		if len(upgradeData["latitude"]) > 0 && len(upgradeData["longitude"]) > 0 {
			progressBar["upgrade_geolocation"] = 1
		}
	}

	// добавлена ли обещанная сумма
	promisedAmount, err := c.Single("SELECT id FROM promised_amount WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	// возможно юзер уже отправил запрос на добавление обещенной суммы
	last_tx, err = c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"NewPromisedAmount"}), 1, c.TimeFormat)
	if (len(last_tx) > 0 && (len(last_tx[0]["queue_tx"]) > 0 || len(last_tx[0]["tx"]) > 0)) || promisedAmount > 0 {
		progressBar["promised_amount"] = 1
	}

	// установлена ли комиссия
	commission, err := c.Single("SELECT commission FROM commission WHERE user_id  =  ?", c.UserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	// возможно юзер уже отправил запрос на добавление комиссии
	last_tx, err = c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangeCommission"}), 1, c.TimeFormat)
	if (len(last_tx) > 0 && (len(last_tx[0]["queue_tx"]) > 0 || len(last_tx[0]["tx"]) > 0)) || len(commission) > 0 {
		progressBar["commission"] = 1
	}

	// голосование за параметры валют. для простоты смотрим в голоса за реф %
	vote, err := c.Single("SELECT user_id FROM votes_referral WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	last_tx, err = c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"VotesComplex"}), 1, c.TimeFormat)
	if (len(last_tx) > 0 && (len(last_tx[0]["queue_tx"]) > 0 || len(last_tx[0]["tx"]) > 0)) || vote > 0 {
		progressBar["vote"] = 1
	}
	if c.SessRestricted == 0 {
		// выполнялись ли задания
		myTasks, err := c.Single("SELECT id FROM " + c.MyPrefix + "my_tasks").Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if myTasks > 0 {
			progressBar["tasks"] = 1
		}
	}

	// сколько майнеров зарегались по ключам данного юзера
	progressBar["referral"], err = c.Single(`
		SELECT count(miner_id)
		FROM users
		LEFT JOIN miners_data on miners_data.user_id = users.user_id
		WHERE referral = ? AND
					 miner_id > 0
	`, c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	// итог
	progressPct := progressBarPct["begin"]
	for name, result := range progressBar {
		if name == "referral" {
			progressPct += progressBarPct[name] * result
		} else {
			progressPct += progressBarPct[name]
		}
	}
	progressBar["begin"] = 1

	log.Debug("ProgressBar end")
	if !c.ContentInc {
		data, err := static.Asset("static/templates/progress_bar.html")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		t := template.Must(template.New("template").Parse(string(data)))
		b := new(bytes.Buffer)
		t.ExecuteTemplate(b, "progressBar", &progressBarPage{Lang: c.Lang, ProgressPct: progressPct})
		return b.String(), nil
	} else {
		TemplateStr, err := makeTemplate("progress", "progress", &progressBarPage{
			Lang:           c.Lang,
			ProgressBar:    progressBar,
			ProgressBarPct: progressBarPct,
			ProgressPct:    progressPct})
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		return TemplateStr, nil
	}
}
