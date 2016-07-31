// emailserv
package main

import (
	"net/http"
	"bytes"
	"fmt"
	"time"
	"archive/zip"
)

func BytesInfoHeader(size int, filename string) (*zip.FileHeader, error) {
	fh := &zip.FileHeader{
		Name:               filename,
		UncompressedSize64: uint64(size),
		UncompressedSize:   uint32(size),
		Method:             zip.Deflate,
	}
	fh.SetModTime(time.Now())
	//   fh.SetMode(fi.Mode())
	return fh, nil
}

func backupHandler(w http.ResponseWriter, r *http.Request) {
	
	if r.URL.Query().Get("test") != GSettings.Password {
		_,_,ok := checkLogin( w, r )
		if !ok {
			return
		}
	}
	
	out := new(bytes.Buffer)
	list,_ := GDB.GetAll(`select user_id, email, verified from users order by user_id`, -1 )
	for _,iout := range list {
		out.WriteString(fmt.Sprintf("%s,%s,%s\r\n", iout[`user_id`], iout[`email`], iout[`verified`] ))	
	}
	buf := new(bytes.Buffer)
	z := zip.NewWriter(buf)
	header, _ := BytesInfoHeader( out.Len(), `dcoin_emails.csv`)
	f,_ := z.CreateHeader(header)
	f.Write(out.Bytes())
	z.Close()
	w.Header().Set(`Content-Description`, `File Transfer`) 
    w.Header().Set(`Content-Type`, `application/zip`)
    w.Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s.zip"`, 
	             time.Now().Format(`060102-150405`)))
	w.Write(buf.Bytes())//out.Bytes())
}
