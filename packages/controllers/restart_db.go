package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/daemons"
	"regexp"
)

func (c *Controller) RestartDb() (string, error) {
	// Обнуляем timeSynchro для Synchronization Blockchain
	timeSynchro = 0
	if ok, _ := regexp.MatchString(`(\:\:)|(127\.0\.0\.1)`, c.r.RemoteAddr); ok {
		err := daemons.ClearDb(nil, "")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	} else {
		return "", utils.ErrInfo("Access denied for "+c.r.RemoteAddr)
	}
	return "", nil
}
