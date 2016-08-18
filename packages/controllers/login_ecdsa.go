package controllers

import (
//	"bytes"
//	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
//	"html/template"
//	"fmt"
)

type loginECDSAPage struct {
	Lang          map[string]string
	Title         string
/*	MyModalIdName string
	UserID        int64
	PoolTechWorks int
	SetupPassword bool
	Community     bool
	Mobile        bool
	SignUp        bool
	Desktop bool*/
}

func (c *Controller) LoginECDSA() (string, error) {

/*	var pool_tech_works int

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
	var setupPassword bool
	if !c.Community {
		// Нельзя зарегистрироваться если в my_table уже есть статус
		if status, err := c.Single("SELECT status FROM my_table").String(); err == nil &&
			status != "waiting_set_new_key" && status != "my_pending" {
			signUp = false
		}
		
		setupPassword_, err := c.Single("SELECT setup_password FROM config").String()
		if err != nil {
			return "", err
		}
		myKey, err := c.GetMyPublicKey(c.MyPrefix)
		if err != nil {
			return "", err
		}
		if len(myKey) == 0 && (len(setupPassword_) > 0 || setupPassword_ == string(utils.DSha256(""))) {
			setupPassword = true
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
		SetupPassword: setupPassword,
		Community:     c.Community,
		SignUp:        signUp,
		Desktop: utils.Desktop(),
		Mobile:        utils.Mobile()})
	if err != nil {
		return "", err
	}
	return b.String(), nil*/
	TemplateStr, err := makeTemplate("login", "loginECDSA", &loginECDSAPage{
		Lang:                  c.Lang,
		Title:                 "Login",		
/*		MyWalletData:          MyWalletData,
		Title:                 "modalAnonym",
*/})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
