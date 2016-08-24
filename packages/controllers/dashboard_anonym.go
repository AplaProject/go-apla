package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)


type dashboardAnonymPage struct {
	Lang                  map[string]string
	Title                 string
	CountSign             int
	Amount                string
	CountSignArr          []int
	SignData              string
	ShowSignData          bool
}

func (c *Controller) DashboardAnonym() (string, error) {
	amount := `0`
	
/*	wallet_id,err := c.GetMyWalletId()
	if err != nil {
		return "", utils.ErrInfo(err)
	}*/
	
	if c.SessWalletId > 0 {
		var err error
		amount,err = c.Single("select amount from dlt_wallets where wallet_id=?", c.SessWalletId ).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	TemplateStr, err := makeTemplate("dashboard_anonym", "dashboardAnonym", &dashboardAnonymPage{
		CountSignArr:          c.CountSignArr,
		CountSign:             c.CountSign,
		Lang:                  c.Lang,
		Title:                 "Home",
		Amount:                amount,
		ShowSignData:          c.ShowSignData,
		SignData:              ""})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
