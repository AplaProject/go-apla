package controllers

import (
	"encoding/json"
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) SaveUserCoords() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()
	coordsJson := c.r.FormValue("coords_json")
	var coords [][2]int
	err := json.Unmarshal([]byte(coordsJson), &coords)
	if err != nil {
		return "", err
	}
	coordsType := c.r.FormValue("type")
	if coordsType != "face" && coordsType != "profile" {
		return "", errors.New("Incorrect type")
	}
	coordsJson_, err := json.Marshal(coords)
	if err != nil {
		return "", err
	}
	err = c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET "+coordsType+"_coords = ?", string(coordsJson_))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if coordsType == "face" {
		if count, _ := utils.DB.GetCountMiners(); count <= 1000 {
			if coordsJson_, err = json.Marshal(coords[:6]); err != nil {
				return "", err
			}
			err = c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET profile_coords = ?", string(coordsJson_))
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
	}
	return `{"success":"ok"}`, nil
}
