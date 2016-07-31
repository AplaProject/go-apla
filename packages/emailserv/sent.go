// emailserv
package main

import (
	"net"
	"net/http"
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func IntToIP(ip uint32) string {
	result := make(net.IP, 4)
	result[3] = byte(ip)
	result[2] = byte(ip >> 8)
	result[1] = byte(ip >> 16)
	result[0] = byte(ip >> 24)
	return result.String()
}

func sentHandler(w http.ResponseWriter, r *http.Request) {
	
	_,_,ok := checkLogin( w, r )
	if !ok {
		return
	}

	
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	list,_ := GDB.GetAll(`select * from log order by id desc`, 50 )
	for i,_ := range list {
		item := list[i]
		ip := utils.StrToInt( item[`ip`])
		if (ip == 1) {
			list[i][`ip`] = `daemon`
		} else {
			list[i][`ip`] = IntToIP( uint32(ip))
		}
		cmds := []string{`UNKNOWN`, `NEW`, `TEST`, `ADMINMSG`, `CASHREQ`,
				`CHANGESTAT`, `DCCAME`, `DCSENT`, `UPDPRIMARY`, `UPDEMAIL`,
				`UPDSMS`, `VOTERES`, `VOTETIME`, `NEWVER`, `NODETIME`, `SIGNUP`, `BALANCE`, 
				`EXREQUEST`, `EXANSWER`, `REFREADY`}
	
		cmd := utils.StrToInt(item[`cmd`]) 
		if cmd < len( cmds ) {
			list[i][`cmd`] = cmds[cmd]
		}
	}
	data[`List`] = list

	if err := GPageTpl.ExecuteTemplate(out, `sent`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
