// emailserv
package main

import (
	"net/http"
	"bytes"
	"fmt"
)

func banHandler(w http.ResponseWriter, r *http.Request) {
	
	_,_,ok := checkLogin( w, r )
	if !ok {
		return
	}
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	r.ParseForm()
	email := r.PostFormValue(`email`)	
	if len(email) > 0 {
		if err:= GDB.ExecSql(`update users set verified=-1 where email=?`, email ); err != nil {
			data[`message`] = err.Error()
		} else if err := GDB.ExecSql(`INSERT INTO stoplist ( email, error, uptime, ip )
				VALUES ( ?, ?, datetime('now'), ? )`, email, `Added by admin`, 0 ); err != nil {
			data[`message`] = err.Error()
		} else {
			data[`message`] = fmt.Sprintf(`Email %s добавлен в стоп-лист`, email )
		}
	} /*else {
		data[`message`] = `Не указан email`
	}*/
	
	if err := GPageTpl.ExecuteTemplate(out, `ban`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
