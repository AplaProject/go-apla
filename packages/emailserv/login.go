// emailserv
package main

import (
	"net/http"
	"bytes"
	"time"
	"crypto/md5"
	"fmt"
)

var attempts  map[uint32]int

func init() {
	attempts = make( map[uint32]int )
}

func SetCookie( name, value string, interval int, response http.ResponseWriter ) {
	var expire time.Time
	if interval > 0 {
		expire = time.Now().AddDate(0, 0, interval )
	} else {
		expire = time.Now().Add( time.Duration( -interval ) * time.Hour )
	}
	cookie := &http.Cookie{Name: name, Value: value, Path: `/`,
			Expires: expire, RawExpires: expire.Format(time.UnixDate)}
    http.SetCookie( response, cookie)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	data := make( map[string]interface{})
	if len(r.FormValue(`password`)) > 0 {
		ip,_ := getIP( r )
		if _, ok := attempts[ip]; ok && attempts[ip] >= 5 {
			data[`message`] = `Blocked`
		} else	if r.FormValue(`password`) == GSettings.Password {
			SetCookie( `admpass`, fmt.Sprintf("%x", md5.Sum([]byte(GSettings.Password))), 90, w )
			http.Redirect(w, r, `/` + GSettings.Admin + `/`, http.StatusFound)
			return
		}
		attempts[ip]++
	}
	data[`Path`] = GSettings.Admin
	out := new(bytes.Buffer)

	if err := GPageTpl.ExecuteTemplate(out, `login`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	
	w.Write(out.Bytes())
}
