// emailserv
package main

import (
	"net/http"
	"strings"
	"hash/crc32"
	"strconv"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func unsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	
	uid := r.URL.Query().Get("uid")
	idemail := strings.Split( uid, `-` )
	if len(idemail)!=2 {
		w.Write( []byte(`Wrong parameters`) )
		return
	} 
	userId := utils.StrToInt64( idemail[0] )
	email,_ := GDB.Single( `select email from users where user_id=?`, userId ).String()
	if crc, err := strconv.ParseUint(idemail[1], 32, 32); err != nil || crc != uint64(crc32.ChecksumIEEE([]byte(email))) {
		w.Write( []byte(`Wrong CRC` ))
		return
	}
	if err :=  GDB.ExecSql(`update users set verified=-2 where email=?`, email ); err != nil {
		w.Write( []byte(`System error` ) )
		return
	}
	ipval,_ := getIP( r )
	if err := GDB.ExecSql(`INSERT INTO stoplist ( email, error, uptime, ip )
				VALUES ( ?, ?, datetime('now'), ? )`, email, `Unsubscribe`, ipval ); err != nil {
	}
	w.Write( []byte(fmt.Sprintf(`Email %s has been successfully unsubscribed`, email )))
}
