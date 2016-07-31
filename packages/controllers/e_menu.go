// e_menu
package controllers

import (
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html/template"
)

type emenuPage struct {
	Lang           map[string]string
	LangInt        int64
	UserId            int64
	EHost             string
	MyWallets      []map[string]string
}

func (c *Controller) EMenu() (string, error) {
	var ( myWallets []map[string]string
		err error
	)
	eHost := c.EConfig["domain"]
		
	if c.SessUserId > 0 {
		myWallets, err = c.getMyWallets()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	funcMap := template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	data, err := static.Asset("static/templates/e_menu.html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	t := template.Must(template.New("template").Funcs(funcMap).Parse(string(data)))
	b := new(bytes.Buffer)	
	err = t.ExecuteTemplate(b, "e_menu", &emenuPage{ Lang: c.Lang, LangInt: c.LangInt, 
		EHost: eHost, UserId: c.SessUserId,	MyWallets: myWallets })
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return b.String(), nil
}
