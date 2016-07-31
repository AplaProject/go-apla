package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type AddCfProjectDataPage struct {
	Alert          string
	SignData       string
	ShowSignData   bool
	Lang           map[string]string
	UserId         int64
	TxType         string
	TxTypeId       int64
	TimeNow        int64
	CountSignArr   []int
	ProjectId      int64
	Id             int64
	CfData         map[string]string
	CfCurrencyName string
	CfLng          map[string]string
}

func (c *Controller) AddCfProjectData() (string, error) {

	var err error

	txType := "CfProjectData"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	projectId := int64(utils.StrToFloat64(c.Parameters["projectId"]))
	id := int64(utils.StrToFloat64(c.Parameters["id"]))
	cfData := make(map[string]string)
	if id > 0 {
		log.Debug("id:", id)
		cfData, err = c.OneRow("SELECT * FROM cf_projects_data WHERE id = ?", id).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		projectId = utils.StrToInt64(cfData["project_id"])
	}

	cfCurrencyName, err := c.Single("SELECT project_currency_name FROM cf_projects WHERE id  =  ?", projectId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	CfLng, err := c.GetAllCfLng()
	log.Debug("CfData", cfData)
	TemplateStr, err := makeTemplate("add_cf_project_data", "addCfProjectData", &AddCfProjectDataPage{
		Alert:          c.Alert,
		Lang:           c.Lang,
		CountSignArr:   c.CountSignArr,
		ShowSignData:   c.ShowSignData,
		UserId:         c.SessUserId,
		TimeNow:        timeNow,
		TxType:         txType,
		TxTypeId:       txTypeId,
		ProjectId:      projectId,
		Id:             id,
		CfData:         cfData,
		CfCurrencyName: cfCurrencyName,
		CfLng:          CfLng})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
