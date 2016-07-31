package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"

	"encoding/json"
	"errors"
	"regexp"
	"time"
)

type CfPagePreviewPage struct {
	SignData             string
	ShowSignData         bool
	TxType               string
	TxTypeId             int64
	TimeNow              int64
	UserId               int64
	Alert                string
	Lang                 map[string]string
	CountSignArr         []int
	CfLng                map[string]string
	CurrencyList         map[int64]string
	CfUrl                string
	ShowHeaders          bool
	Page                 string
	CfCurrencyName       string
	LangId               int64
	ProjectId            int64
	BlurbImg             string
	HeadImg              string
	DescriptionImg       string
	Picture              string
	VideoType            string
	VideoUrlId           string
	NewsImg              string
	Links                [][]string
	ImgBlank             string
	Project              map[string]string
	ProjectLang          map[string]string
	ProjectFunding       float64
	ProjectCountFunders  int64
	ProjectFunders       []map[string]string
	ProjectComments      []map[string]string
	LangComments         map[string]string
	ProjectCountComments int64
	AuthorInfo           map[string]string
	PagesArray           []string
	ConfigCfPs           map[string][]string
	ProjectPs            map[string]string
	Black                int64
}

func (c *Controller) CfPagePreview() (string, error) {

	var err error

	txType := "CfComment"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	cfUrl, err := c.GetCfUrl()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	showHeaders := false
	if len(c.r.FormValue("blurb_img")) > 0 {
		showHeaders = true
	}

	page := c.Parameters["page"]
	if len(page) > 0 {
		if ok, _ := regexp.MatchString(`^(?i)[a-z]{0,10}$`, page); !ok {
			return "", errors.New("incorrect page")
		}
	}
	cfCurrencyName := c.Parameters["onlyCfCurrencyName"]
	if len(page) == 0 {
		page = "home"
	}

	langId := int64(utils.StrToFloat64(c.Parameters["lang_id"]))
	projectId := int64(utils.StrToFloat64(c.Parameters["onlyProjectId"]))
	log.Debug("projectId:", projectId)
	var blurbImg, headImg, descriptionImg, picture, videoType, videoUrlId, newsImg string
	imgBlank := cfUrl + "static/img/blank.png"

	var links [][]string
	if projectId > 0 || len(cfCurrencyName) > 0 {
		if projectId == 0 {
			projectId, err = c.Single("SELECT id FROM cf_projects WHERE project_currency_name = ?", cfCurrencyName).Int64()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
		data := make(map[string]string)
		if langId > 0 {
			data, err = c.OneRow("SELECT * FROM cf_projects_data WHERE project_id = ? AND lang_id = ?", projectId, langId).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		} else { // Если язык не указан, то просто берем первое добавленное описание
			data, err = c.OneRow("SELECT * FROM cf_projects_data WHERE project_id = ? ORDER BY id ASC", projectId).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			langId = utils.StrToInt64(data["lang_id"])
		}
		blurbImg = data["blurb_img"]
		headImg = data["head_img"]
		descriptionImg = data["description_img"]
		picture = data["picture"]
		videoType = data["video_type"]
		videoUrlId = data["video_url_id"]
		newsImg = data["news_img"]
		if len(data["links"]) > 0 && data["links"] != "0" {
			var links_ [][]interface{}
			err = json.Unmarshal([]byte(data["links"]), &links_)
			if err != nil {
				log.Debug("data links:", data["links"])
				return "", utils.ErrInfo(err)
			}
			for _, v := range links_ {
				var l []string
				for _, v2 := range v {
					str := utils.InterfaceToStr(v2)
					if len(str) == 0 {
						return "", utils.ErrInfo(errors.New("Incorrect links"))
					}
					l = append(l, str)
				}
				links = append(links, l)
			}
		}
	} else {
		log.Debug("FormValue", c.r.Form)

		blurbImg = c.r.FormValue("blurb_img")
		headImg = c.r.FormValue("head_img")
		descriptionImg = c.r.FormValue("description_img")
		picture = c.r.FormValue("blurb_img")
		videoType = c.r.FormValue("video_type")
		videoUrlId = c.r.FormValue("video_url_id")
		newsImg = c.r.FormValue("news_img")

		if !utils.CheckInputData(c.r.FormValue("project_id"), "int") {
			return "", errors.New("Incorrect project_id")
		}
		if !utils.CheckInputData(blurbImg, "img_url") {
			blurbImg = imgBlank
		}
		if !utils.CheckInputData(headImg, "img_url") {
			headImg = imgBlank
		}
		if !utils.CheckInputData(descriptionImg, "img_url") {
			descriptionImg = imgBlank
		}
		if !utils.CheckInputData(picture, "img_url") {
			picture = imgBlank
		}
		if !utils.CheckInputData(newsImg, "img_url") {
			newsImg = imgBlank
		}
		if !utils.CheckInputData(videoType, "video_type") {
			videoType = ""
		}
		if !utils.CheckInputData(videoUrlId, "video_url_id") {
			videoUrlId = ""
		}

		if len(c.Parameters["links"]) > 0 {
			var links_ [][]interface{}
			err = json.Unmarshal([]byte(c.Parameters["links"]), &links_)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			for _, v := range links_ {
				var l []string
				for _, v2 := range v {
					str := utils.InterfaceToStr(v2)
					if len(str) == 0 {
						return "", utils.ErrInfo(errors.New("Incorrect links"))
					}
					l = append(l, str)
				}
				links = append(links, l)
			}
		}
	}

	project, err := c.OneRow("SELECT * FROM cf_projects WHERE id = ?", projectId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	// сколько дней осталось
	days := utils.Round((utils.StrToFloat64(project["end_time"])-float64(utils.Time()))/86400, 0)
	if days <= 0 {
		days = 0
	}
	project["days"] = utils.Float64ToStr(days)
	if project["close_block_id"] != "0" || project["del_block_id"] != "0" {
		project["ended"] = "1"
	} else {
		project["ended"] = "0"
	}

	// дата старта
	t := time.Unix(utils.StrToInt64(project["start_time"]), 0)
	project["start_date"] = t.Format(c.TimeFormat)

	// в какой валюте идет сбор
	project["currency"] = c.CurrencyList[utils.StrToInt64(project["currency_id"])]

	// на каких языках есть описание
	// для home/news можно скрыть язык
	addSql := ""
	if page == "home" || page == "news" {
		addSql = " AND hide = 0 "
	}
	projectLang, err := c.GetMap(`SELECT id, lang_id FROM cf_projects_data WHERE project_id = ? `+addSql, "id", "lang_id", projectId)

	// сколько собрано средств
	projectFunding, err := c.Single("SELECT sum(amount) FROM cf_funding WHERE project_id  =  ? AND del_block_id  =  0", projectId).Float64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// сколько всего фундеров
	projectCountFunders, err := c.Single("SELECT count(id) FROM cf_funding WHERE project_id  =  ? AND del_block_id  =  0 GROUP BY user_id", projectId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// список фундеров
	var q string
	if c.ConfigIni["db_type"] == "postgresql" {
		q = `SELECT DISTINCT cf_funding.user_id, sum(amount) as amount, time,  name, avatar
			FROM cf_funding
			LEFT JOIN users ON users.user_id = cf_funding.user_id
			WHERE project_id = ? AND
						del_block_id = 0
			GROUP BY cf_funding.user_id, time, name, avatar
			ORDER BY time DESC
			LIMIT 100`
	} else {
		q = `SELECT cf_funding.user_id, sum(amount) as amount, time,  name, avatar
			FROM cf_funding
			LEFT JOIN users ON users.user_id = cf_funding.user_id
			WHERE project_id = ? AND
						del_block_id = 0
			GROUP BY cf_funding.user_id
			ORDER BY time DESC
			LIMIT 100`
	}
	projectFunders, err := c.GetAll(q, 100, projectId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for k, v := range projectFunders {
		t := time.Unix(utils.StrToInt64(v["time"]), 0)
		projectFunders[k]["time"] = t.Format(c.TimeFormat)
		if len(v["avatar"]) == 0 {
			projectFunders[k]["avatar"] = cfUrl + "static/img/noavatar.png"
		}
		if len(v["name"]) == 0 {
			projectFunders[k]["name"] = "Noname"
		}
	}

	// список комментов
	var projectComments []map[string]string
	if langId > 0 {
		projectComments, err = c.GetAll(`
				SELECT users.user_id, comment, time, name, avatar
				FROM cf_comments
				LEFT JOIN users ON users.user_id = cf_comments.user_id
				WHERE project_id = ? AND
							 lang_id = ?
				ORDER BY time DESC
				LIMIT 100
				`, 100, projectId, langId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		for k, v := range projectComments {
			t := time.Unix(utils.StrToInt64(v["time"]), 0)
			projectComments[k]["time"] = t.Format(c.TimeFormat)
			if len(v["avatar"]) == 0 {
				projectComments[k]["avatar"] = cfUrl + "static/img/noavatar.png"
			}
			if len(v["name"]) == 0 {
				projectComments[k]["name"] = "Noname"
			}
		}
	}
	// сколько всего комментов на каждом языке
	if c.ConfigIni["db_type"] == "postgresql" {
		q = `SELECT DISTINCT lang_id, count(id) as count FROM cf_comments WHERE project_id = ? GROUP BY lang_id`
	} else {
		q = `SELECT lang_id, count(id) as count FROM cf_comments WHERE project_id = ?`
	}
	langComments, err := c.GetMap(q, "lang_id", "count", projectId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	var projectCountComments int64
	for _, v := range langComments {
		projectCountComments += utils.StrToInt64(v)
	}

	cfLng, err := c.GetAllCfLng()

	// инфа об авторе проекта
	authorInfo, err := c.GetCfAuthorInfo(project["user_id"], cfUrl)

	// возможно наш юзер фундер
	project["funder"] = ""
	if c.SessUserId > 0 {
		project["funder"], err = c.Single("SELECT id FROM cf_funding WHERE project_id  =  ? AND user_id  =  ? AND del_block_id  =  0", projectId, c.SessUserId).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	pagesArray := []string{"home", "news", "funders", "comments"}
	configCfPs := make(map[string][]string)
	if len(c.NodeConfig["cf_ps"]) > 0 {
		//{"1":["Credit card"],"2":["ik","MTS, Megafon, W1, Paxum"],"3":["pm","Perfect Money"]}
		err = json.Unmarshal([]byte(c.NodeConfig["cf_ps"]), &configCfPs)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	// узнаем, какие платежные системы доступны данному проекту
	projectPs, err := c.OneRow("SELECT * FROM cf_projects_ps WHERE project_id  =  ?", projectId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// узнаем, не в блек-листе ли проект
	black, err := c.Single("SELECT project_id FROM cf_blacklist WHERE project_id  =  ?", projectId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if black > 0 {
		return "", errors.New("Black project")
	}

	TemplateStr, err := makeTemplate("cf_page_preview", "cfPagePreview", &CfPagePreviewPage{
		Alert:                c.Alert,
		Lang:                 c.Lang,
		CountSignArr:         c.CountSignArr,
		ShowSignData:         c.ShowSignData,
		UserId:               c.SessUserId,
		TimeNow:              timeNow,
		TxType:               txType,
		TxTypeId:             txTypeId,
		SignData:             "",
		CfLng:                cfLng,
		CurrencyList:         c.CurrencyList,
		CfUrl:                cfUrl,
		ShowHeaders:          showHeaders,
		Page:                 page,
		CfCurrencyName:       cfCurrencyName,
		LangId:               langId,
		ProjectId:            projectId,
		BlurbImg:             blurbImg,
		HeadImg:              headImg,
		DescriptionImg:       descriptionImg,
		Picture:              picture,
		VideoType:            videoType,
		VideoUrlId:           videoUrlId,
		NewsImg:              newsImg,
		Links:                links,
		ImgBlank:             imgBlank,
		Project:              project,
		ProjectLang:          projectLang,
		ProjectFunding:       projectFunding,
		ProjectCountFunders:  projectCountFunders,
		ProjectFunders:       projectFunders,
		ProjectComments:      projectComments,
		LangComments:         langComments,
		ProjectCountComments: projectCountComments,
		AuthorInfo:           authorInfo,
		PagesArray:           pagesArray,
		ConfigCfPs:           configCfPs,
		ProjectPs:            projectPs,
		Black:                black})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
