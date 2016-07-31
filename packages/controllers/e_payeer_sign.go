package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
	//"encoding/base64"
	"fmt"
)

func (c *Controller) EPayeerSign() (string, error) {

	c.r.ParseForm()
	sign := strings.ToUpper(string(utils.Sha256(c.EConfig["payeer_id"] + ":" + c.r.FormValue("m_orderid") + ":" + c.r.FormValue("m_amount") + ":USD:" + c.r.FormValue("m_desc") + ":" + c.EConfig["payeer_s_key"])))
	fmt.Println(sign)
	fmt.Println(c.EConfig["payeer_id"] + ":" + c.r.FormValue("m_orderid") + ":" + c.r.FormValue("m_amount") + ":USD:" + c.r.FormValue("m_desc") + ":" + c.EConfig["payeer_s_key"])
	return sign, nil
}
