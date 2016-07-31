package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type bugReportingPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Lang         map[string]string
	UserId       int64
	ParentId     int64
	Messages     []map[string]string
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) BugReporting() (string, error) {

	txType := "MessageToAdmin"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	// если юзер тыкнул по какой-то ветке сообщений, то тут будет parent_id, т.е. id этой ветки
	parentId := int64(utils.StrToFloat64(c.Parameters["parent_id"]))
	messages, err := c.GetAll(`
			SELECT *
			FROM  `+c.MyPrefix+`my_admin_messages
			WHERE message_type = 0 AND
						 (parent_id = ? OR id = ?)
			ORDER BY id DESC
			`, -1, parentId, parentId)

	TemplateStr, err := makeTemplate("bug_reporting", "bugReporting", &bugReportingPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		UserId:       c.SessUserId,
		ParentId:     parentId,
		Messages:     messages,
		CountSignArr: c.CountSignArr,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
