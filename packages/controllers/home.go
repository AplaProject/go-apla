package controllers

func (c *Controller) Home() (string, error) {

/*	TemplateStr, err := makeTemplate("dashboard_anonym", "dashboardAnonym", &dashboardAnonymPage{
		CountSignArr:          c.CountSignArr,
		CountSign:             c.CountSign,
		Lang:                  c.Lang,
		Title:                 "Home",
		ShowSignData:          c.ShowSignData,
		SignData:              ""})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil*/
	return c.DashboardAnonym()
}
