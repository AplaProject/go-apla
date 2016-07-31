// emailserv
package main

import (
	"net/http"
	"bytes"
	//"fmt"
	//"github.com/DayLightProject/go-daylight/packages/consts"

)

func adminHandler(w http.ResponseWriter, r *http.Request) {
	
	_, _,ok := checkLogin( w, r )
	if !ok {
		return
	}
/*	for i, name := range consts.Countries {
		fmt.Println( i, name )
	}*/
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	data[`Path`] = GSettings.Admin
	if err := GPageTpl.ExecuteTemplate(out, `admin`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
