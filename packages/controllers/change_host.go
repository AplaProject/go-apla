package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

type changeHostPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Data         map[string]string
	Lang         map[string]string
	UserId       int64
	LimitsText   string
	Community    bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) ChangeHost() (string, error) {

	txType := "ChangeHost"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	if !c.PoolAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("access denied"))
	}

	data, err := c.OneRow("SELECT http_host, tcp_host, host_status FROM " + c.MyPrefix + "my_table").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	data2, err := c.OneRow("SELECT http_host, tcp_host, e_host FROM miners_data WHERE user_id = ?", c.SessUserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(data["http_host"]) == 0 {
		data["http_host"] = data2["http_host"]
	}
	if len(data["tcp_host"]) == 0 {
		data["tcp_host"] = data2["tcp_host"]
	}
	if len(data["e_host"]) == 0 {
		data["e_host"] = data2["e_host"]
	}
	if data["e_host"] == "" {
		data["e_host"] = "0"
	}

	statusArray := map[string]string{"my_pending": c.Lang["local_pending"], "approved": c.Lang["status_approved"]}
	data["host_status"] = statusArray[data["host_status"]]

	limitsText := strings.Replace(c.Lang["change_host_limits_text"], "[limit]", utils.Int64ToStr(c.Variables.Int64["limit_change_host"]), -1)
	limitsText = strings.Replace(limitsText, "[period]", c.Periods[c.Variables.Int64["limit_change_host_period"]], -1)

	TemplateStr, err := makeTemplate("change_host", "changeHost", &changeHostPage{
		Alert:        c.Alert,
		UserId:       c.SessUserId,
		CountSignArr: c.CountSignArr,
		Data:         data,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		LimitsText:   limitsText,
		ShowSignData: c.ShowSignData,
		Community:    c.Community,
		SignData:     "",
		Lang:         c.Lang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
