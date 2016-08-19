package controllers

import (
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html/template"
)

type loginStruct struct {
	Lang          map[string]string
	MyModalIdName string
	UserID        int64
	PoolTechWorks int
	Community     bool
	Mobile        bool
	SignUp        bool
	Desktop bool
}

func (c *Controller) Login() (string, error) {

	var pool_tech_works int

	funcMap := template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	data, err := static.Asset("static/templates/login.html")
	if err != nil {
		return "", err
	}
	modal, err := static.Asset("static/templates/modal.html")
	if err != nil {
		return "", err
	}

	t := template.Must(template.New("template").Funcs(funcMap).Parse(string(data)))
	t = template.Must(t.Parse(string(modal)))

	b := new(bytes.Buffer)
	signUp := true
	// есть ли установочный пароль и был ли начально записан ключ
	if !c.Community {
		// Нельзя зарегистрироваться если в my_table уже есть статус
		if status, err := c.Single("SELECT status FROM my_table").String(); err == nil &&
			status != "waiting_set_new_key" && status != "my_pending" {
			signUp = false
		}
		
	}
	//fmt.Println(c.Lang)
	// проверим, не идут ли тех. работы на пуле
	if len(c.NodeConfig["pool_admin_user_id"]) > 0 && c.NodeConfig["pool_admin_user_id"] != utils.Int64ToStr(c.UserId) && c.NodeConfig["pool_tech_works"] == "1" && c.Community {
		pool_tech_works = 1
	} else {
		pool_tech_works = 0
	}
	err = t.ExecuteTemplate(b, "login", &loginStruct{
		Lang:          c.Lang,
		MyModalIdName: "myModalLogin",
		UserID:        c.UserId,
		PoolTechWorks: pool_tech_works,
		Community:     c.Community,
		SignUp:        signUp,
		Desktop: utils.Desktop(),
		Mobile:        utils.Mobile()})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
