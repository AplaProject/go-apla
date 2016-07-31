// emailserv
package main

import (
	"net/http"
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/utils"
)


func notificationsHandler(w http.ResponseWriter, r *http.Request) {
	_,_,ok := checkLogin( w, r )
	if !ok {
		return
	}
	
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	list,_ := utils.DB.GetAll(`select * from notifications order by id desc`, 50 )
	for i,_ := range list {
		item := list[i]
		cmds := []string{`UNKNOWN`, `NEW`, `TEST`, `ADMINMSG`, `CASHREQ`,
				`CHANGESTAT`, `DCCAME`, `DCSENT`, `UPDPRIMARY`, `UPDEMAIL`,
				`UPDSMS`, `VOTERES`, `VOTETIME`, `NEWVER`, `NODETIME`, `SIGNUP`, `BALANCE`, 
				`EXREQUEST`, `EXANSWER`, `REFREADY`}
	
		cmd := utils.StrToInt(item[`cmd_id`]) 
		if cmd < len( cmds ) {
			list[i][`cmd_id`] = cmds[cmd]
		}
	}
	data[`List`] = list

	if err := GPageTpl.ExecuteTemplate(out, `notifications`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
