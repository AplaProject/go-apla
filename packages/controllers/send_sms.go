package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io/ioutil"
	"net/http"
)

func (c *Controller) SendSms() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()
	text := c.r.FormValue("text")

	sms_http_get_request, err := c.Single("SELECT sms_http_get_request FROM " + c.MyPrefix + "my_table").String()
	if err != nil {
		result, _ := json.Marshal(map[string]string{"error": fmt.Sprintf(`%s`, err)})
		return string(result), nil
	}
	resp, err := http.Get(sms_http_get_request + text)
	if err != nil {
		result, _ := json.Marshal(map[string]string{"error": fmt.Sprintf(`%s`, err)})
		return string(result), nil
	}
	defer resp.Body.Close()
	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result, _ := json.Marshal(map[string]string{"error": fmt.Sprintf(`%s`, err)})
		return string(result), nil
	}
	result, _ := json.Marshal(map[string]string{"success": string(htmlData)})
	return string(result), nil

}
