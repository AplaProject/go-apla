// emailserv
package main

import (
	"net/http"
	"bytes"
	"io/ioutil"
	"path/filepath"
)

func newHandler(w http.ResponseWriter, r *http.Request) {
	
	_,_,ok := checkLogin( w, r )
	if !ok {
		return
	}
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	r.ParseForm()
	if len(r.PostFormValue(`send`)) > 0 {
		filename := r.PostFormValue(`filename`)	
		content := r.PostFormValue(`content`)	
		if len( filename ) == 0 {
			data[`message`] = `Не указано имя файла`
		} else if len( content ) == 0 {
			data[`message`] = `Не указан текст`
		} else if err:=ioutil.WriteFile( filepath.Join( `pattern`, filename ), []byte(content), 0644); err != nil {
			data[`message`] = err.Error()
		} else {
			http.Redirect(w, r, `/` + GSettings.Admin + `/patterns`, http.StatusFound)
		}
	}

	if err := GPageTpl.ExecuteTemplate(out, `new`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
