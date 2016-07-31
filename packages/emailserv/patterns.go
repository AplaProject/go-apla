// emailserv
package main

import (
	"net/http"
	"bytes"
	"io/ioutil"
)

func patternsHandler(w http.ResponseWriter, r *http.Request) {
	
	_,_,ok := checkLogin( w, r )
	if !ok {
		return
	}
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	r.ParseForm()
	list := make( []string, 0 )
	files, err := ioutil.ReadDir( `pattern` )
	if err != nil {
		data[`message`] = err.Error()
	} else {
		for _, file := range files {
			list = append( list, file.Name())
		}
	}	
	data[`Items`] = list
	data[`Path`] = GSettings.Admin
//	email := r.PostFormValue(`email`)	
	
	if err := GPageTpl.ExecuteTemplate(out, `patterns`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
