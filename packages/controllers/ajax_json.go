package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"net/http"
	//	"regexp"
	"fmt"
	//	"encoding/json"
)

type answerJson struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Result  bool   `json:"result"`
	Data    string `json:"data"`
}

func AjaxJson(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("ajax_json Recovered", r)
			fmt.Println("ajax_json Recovered", r)
		}
	}()
	c := new(Controller)
	c.r = r
	c.w = w
	c.DCDB = utils.DB
	c.Parameters, _ = c.GetParameters()
	lang := GetLang(w, r, c.Parameters)
	c.Lang = globalLangReadOnly[lang]
	c.LangInt = int64(lang)
	c.Variables,_ = c.GetAllVariables()

	r.ParseForm()
	controllerName := r.FormValue("controllerName")

	answer := []byte(`{"success": false, "error": "", "result": false, "data": ""}`)

	if utils.InSliceString( controllerName, []string{ `CheckHash`, `CheckForm`,`CheckPromised`,`NotifyCounter`, `SendKey` }) {
		if ret, err := CallController(c, controllerName); err == nil {
			answer = []byte(ret)
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(answer)
}
