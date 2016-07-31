// emailserv
package main

import (
	"net/http"
	"bytes"
	"io/ioutil"
	"path/filepath"
)

func editHandler(w http.ResponseWriter, r *http.Request) {
	
	_,_,ok := checkLogin( w, r )
	if !ok {
		return
	}
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	pattern := r.URL.Query().Get("pattern")
	content := ``
	if len(pattern) == 0 {
		data[`message`] = `Не указан шаблон`
	} else {
		r.ParseForm()
		if len(r.PostFormValue(`send`)) > 0 {
			content := r.PostFormValue(`content`)	
			if len( content ) == 0 {
				data[`message`] = `Не указан текст`
			} else if err:=ioutil.WriteFile( filepath.Join( `pattern`, pattern ), []byte(content), 0644); err != nil {
				data[`message`] = err.Error()
			} else {
				data[`message`] = `Текст успешно сохранен`
			}
		}
		if text, err := ioutil.ReadFile( filepath.Join(`pattern`, pattern )); err == nil {
			content = string( text )
		} else {
			content = err.Error()
		}
		
	}
	data[`Filename`] = pattern
	data[`Content`] = content
	
	if err := GPageTpl.ExecuteTemplate(out, `edit`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
