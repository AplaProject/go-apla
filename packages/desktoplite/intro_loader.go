// intro_loader
package main

import (
	"fmt"
	"net/http"
	"html/template"
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/static"	
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/astaxie/beego/config"

)

type introData struct {
	Lang             map[string]string
	PoolUrl          string
}

var (
	introInit  bool 
	globalLangReadOnly map[int]map[string]string
)

func introLoad() {
	globalLangReadOnly = make(map[int]map[string]string)
	for _, v := range consts.LangMap {
		data, err := static.Asset(fmt.Sprintf("static/lang/%d.ini", v))
		if err != nil {
			fmt.Println( err )
		}
		iniconf_, err := config.NewConfigData("ini", []byte(data))
		if err != nil {
			fmt.Println( err )
		}
		//fmt.Println(iniconf_)
		iniconf, err := iniconf_.GetSection("default")
		globalLangReadOnly[v] = make(map[string]string)
		globalLangReadOnly[v] = iniconf
	}
}

func introLoader(w http.ResponseWriter, r *http.Request) {
	if !introInit {
		introLoad()
	}
	alert_success, err := static.Asset("static/templates/alert_success.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	data, err := static.Asset("static/templates/desktoplite.html")
	if err != nil {
		return
	}
	t := template.Must(template.New("template").Parse(string(data)))
	t = template.Must(t.Parse(string(alert_success)))	
	b := new(bytes.Buffer)
	
	idata := introData{ PoolUrl: GETPOOLURL }
	idata.Lang = globalLangReadOnly[1]
	err = t.ExecuteTemplate(b, `DesktopLite`, idata )
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Fprint(w, b.String())
}
