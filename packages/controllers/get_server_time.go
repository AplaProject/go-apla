package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) GetServerTime() (string, error) {
	return `{"time":"`+utils.Int64ToStr(utils.Time())+`"}`, nil
}