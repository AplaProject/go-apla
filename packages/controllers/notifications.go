package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type notificationsPage struct {
	SignData        string
	ShowSignData    bool
	Alert           string
	Lang            map[string]string
	CountSignArr    []int
	MyNotifications map[string]map[string]string
	LangInt         int64
	NodeAdmin       bool
	Data            map[string]string
	Mobile          bool
}

func (c *Controller) Notifications() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	var err error
	data, err := c.OneRow(`
			SELECT email,
						 sms_http_get_request,
						 use_smtp,
						 smtp_server,
						 smtp_port,
						 smtp_ssl,
						 smtp_auth,
						 smtp_username,
						 smtp_password
			FROM ` + c.MyPrefix + `my_table
			`).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	myNotifications := make(map[string]map[string]string)
	myNotifications_, err := c.GetAll("SELECT * FROM "+c.MyPrefix+"my_notifications WHERE name != 'new_version' ORDER BY sort ASC", -1)
	for _, data := range myNotifications_ {
		myNotifications[data["name"]] = map[string]string{"mobile": data["mobile"], "email": data["email"], "sms": data["sms"], "important": data["important"]}
	}
	log.Debug("myNotifications", myNotifications)

	TemplateStr, err := makeTemplate("notifications", "notifications", &notificationsPage{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		SignData:        "",
		MyNotifications: myNotifications,
		NodeAdmin:       c.NodeAdmin,
		LangInt:         c.LangInt,
		Mobile:          utils.Mobile(),
		Data:            data})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
