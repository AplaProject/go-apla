package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type cfCatalogPage struct {
	Lang         map[string]string
	CfUrl        string
	CfCategory   []map[string]string
	CurrencyList map[int64]string
	CurCategory  string
	Projects     map[string]map[string]string
	UserId       int64
	CategoryId   string
}

func (c *Controller) CfCatalog() (string, error) {

	var err error
	log.Debug("CfCatalog")

	categoryId := utils.Int64ToStr(int64(utils.StrToFloat64(c.Parameters["category_id"])))
	log.Debug("categoryId", categoryId)
	var curCategory string
	addSql := ""
	if categoryId != "0" {
		addSql = `AND category_id = ` + categoryId
		curCategory = c.Lang["cf_category_"+categoryId]
	}

	cfUrl := ""

	projects := make(map[string]map[string]string)
	cfProjects, err := c.GetAll(`
			SELECT cf_projects.id, lang_id, blurb_img, country, city, currency_id, end_time, amount
			FROM cf_projects
			LEFT JOIN cf_projects_data ON  cf_projects_data.project_id = cf_projects.id
			WHERE del_block_id = 0 AND
						 end_time > ? AND
						 lang_id = ?
						`+addSql+`
			ORDER BY funders DESC
			LIMIT 100
			`, 100, utils.Time(), c.LangInt)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for _, data := range cfProjects {
		CfProjectData, err := c.GetCfProjectData(utils.StrToInt64(data["id"]), utils.StrToInt64(data["end_time"]), c.LangInt, utils.StrToFloat64(data["amount"]), cfUrl)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		for k, v := range CfProjectData {
			data[k] = v
		}
		projects[data["id"]] = data
	}

	cfCategory := utils.MakeCfCategories(c.Lang)

	TemplateStr, err := makeTemplate("cf_catalog", "cfCatalog", &cfCatalogPage{
		Lang:         c.Lang,
		CfCategory:   cfCategory,
		CurrencyList: c.CurrencyList,
		CurCategory:  curCategory,
		Projects:     projects,
		UserId:       c.SessUserId,
		CategoryId:   categoryId,
		CfUrl:        cfUrl})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
