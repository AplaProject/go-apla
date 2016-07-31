package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type cfProjectChangeCategoryPage struct {
	Alert               string
	SignData            string
	ShowSignData        bool
	Lang                map[string]string
	UserId              int64
	TxType              string
	TxTypeId            int64
	TimeNow             int64
	CountSignArr        []int
	ProjectId           int64
	CategoryId          string
	CfCategory          []map[string]string
	ProjectCurrencyName string
}

func (c *Controller) CfProjectChangeCategory() (string, error) {

	var err error

	txType := "CfProjectData"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	projectId := int64(utils.StrToFloat64(c.Parameters["project_id"]))
	data, err := c.OneRow("SELECT category_id, project_currency_name FROM cf_projects WHERE id= ?", projectId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	categoryId := data["category_id"]
	projectCurrencyName := data["project_currency_name"]

	cfCategory := utils.MakeCfCategories(c.Lang)

	TemplateStr, err := makeTemplate("cf_project_change_category", "cfProjectChangeCategory", &cfProjectChangeCategoryPage{
		Alert:               c.Alert,
		Lang:                c.Lang,
		CountSignArr:        c.CountSignArr,
		ShowSignData:        c.ShowSignData,
		SignData:            fmt.Sprintf(`%d,%d,%d,%d`, txTypeId, timeNow, c.SessUserId, categoryId),
		UserId:              c.SessUserId,
		TimeNow:             timeNow,
		TxType:              txType,
		TxTypeId:            txTypeId,
		ProjectId:           projectId,
		CategoryId:          categoryId,
		CfCategory:          cfCategory,
		ProjectCurrencyName: projectCurrencyName})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
