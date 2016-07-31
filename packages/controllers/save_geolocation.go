package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
	geo "github.com/DayLightProject/go-daylight/packages/geolocation"
	l "log"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

func (c *Controller) SaveGeolocation() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()
	geolocation := c.r.FormValue("geolocation")
	if len(geolocation) > 0 {
		x := strings.Split(geolocation, ", ")
		if len(x) == 2 {
			geolocationLat := utils.Round(utils.StrToFloat64(x[0]), 5)
			geolocationLon := utils.Round(utils.StrToFloat64(x[1]), 5)

			resp, err := geo.GetInfo(geolocationLat, geolocationLon)
			if err != nil {

			}
			if len(resp.Results) > 0 {
				country := resp.GetCountryName()
				l.Println("Country name:", country)
				for i, v := range consts.Countries {
					if v == country {
						l.Println("Country id:", i)
						err := c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET country = ?", i)
						if err != nil {
							return "", utils.ErrInfo(err)
						}
						break
					}
				}
			}
			err = c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET geolocation = ?", utils.Float64ToStrGeo(geolocationLat)+", "+utils.Float64ToStrGeo(geolocationLon))
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
	}
	return `{"error":"0"}`, nil
}
