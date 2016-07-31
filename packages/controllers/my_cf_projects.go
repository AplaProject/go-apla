package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type MyCfProjectsPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	CountSignArr []int
	Lang         map[string]string
	CfLng        map[string]string
	CurrencyList map[int64]string
	Projects     map[string]map[string]string
	UserId       int64
	ProjectsLang map[string]map[string]string
}

func (c *Controller) MyCfProjects() (string, error) {

	var err error

	txType := "NewCfProject"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	projectsLang := make(map[string]map[string]string)
	projects := make(map[string]map[string]string)
	cfProjects, err := c.GetAll(`
			SELECT id, category_id, project_currency_name, country, city, currency_id, end_time, amount
			FROM cf_projects
			WHERE user_id = ? AND del_block_id = 0
			`, -1, c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for _, data := range cfProjects {
		CfProjectData, err := c.GetCfProjectData(utils.StrToInt64(data["id"]), utils.StrToInt64(data["end_time"]), c.LangInt, utils.StrToFloat64(data["amount"]), "")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		for k, v := range CfProjectData {
			data[k] = v
		}
		projects[data["id"]] = data
		lang, err := c.GetMap(`SELECT id, lang_id FROM cf_projects_data WHERE project_id = ?`, "id", "lang_id", data["id"])
		projectsLang[data["id"]] = lang
	}

	cfLng, err := c.GetAllCfLng()

	TemplateStr, err := makeTemplate("my_cf_projects", "myCfProjects", &MyCfProjectsPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		CfLng:        cfLng,
		CurrencyList: c.CurrencyList,
		Projects:     projects,
		ProjectsLang: projectsLang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
