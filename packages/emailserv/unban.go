// emailserv
package main

import (
	"net/http"
	"bytes"
	"fmt"
)

func unbanHandler(w http.ResponseWriter, r *http.Request) {
	
	_,_,ok := checkLogin( w, r )
	if !ok {
		return
	}
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	r.ParseForm()
	email := r.PostFormValue(`email`)	
	if len(email) > 0 {
		if err:= GDB.ExecSql(`update users set verified=0 where email=?`, email ); err != nil {
			data[`message`] = err.Error()
		} else if err:= GDB.ExecSql(`delete from stoplist where email=?`, email ); err != nil {
			data[`message`] = err.Error()
		} else {
			data[`message`] = fmt.Sprintf(`Email %s убран из стоп-листа`, email )
		}
	} else {
		data[`message`] = `Не указан email`
	}
	
	if err := GPageTpl.ExecuteTemplate(out, `unban`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
