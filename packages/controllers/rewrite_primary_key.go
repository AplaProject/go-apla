package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type rewritePrimaryKeyPage struct {
	Alert string
	Lang  map[string]string
}

func (c *Controller) RewritePrimaryKey() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	TemplateStr, err := makeTemplate("rewrite_primary_key", "rewritePrimaryKey", &rewritePrimaryKeyPage{
		Alert: c.Alert,
		Lang:  c.Lang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
