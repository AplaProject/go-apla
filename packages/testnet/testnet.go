// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/go-bindata-assetfs"
	//	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

type Settings struct {
	Port uint32 `json:"port"`
	Path string `json:"path"`
}

type IndexData struct {
}

var (
	GSettings Settings

/*	GDB       *utils.DCDB
 */
)

func FileAsset(name string) ([]byte, error) {
	return static.Asset(name)
}

func getIP(r *http.Request) (uint32, string) {
	var ipval uint32

	remoteAddr := r.RemoteAddr
	var ip string
	if ip = r.Header.Get(XRealIP); len(ip) > 6 {
		remoteAddr = ip
	} else if ip = r.Header.Get(XForwardedFor); len(ip) > 6 {
		remoteAddr = ip
	}
	if strings.Contains(remoteAddr, ":") {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}
	if ipb := net.ParseIP(remoteAddr).To4(); ipb != nil {
		ipval = uint32(ipb[3]) | (uint32(ipb[2]) << 8) |
			(uint32(ipb[1]) << 16) | (uint32(ipb[0]) << 24)
	}
	return ipval, remoteAddr
}

func testnetHandler(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	data, err := static.Asset("static/testnet.html")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
	}
	t, err := template.New("template").Funcs(funcMap).Parse(string(data))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
	}
	b := new(bytes.Buffer)
	err = t.Execute(b, nil)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
	}
	w.Write(b.Bytes())
	return
}

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(`Dir`, err)
	}
	//	os.Chdir(dir)
	logfile, err := os.OpenFile(filepath.Join(dir, "testnet.log"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(`Testnet log`, err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)
	params, err := ioutil.ReadFile(filepath.Join(dir, `settings.json`))
	if err != nil {
		log.Fatalln(dir, `Settings.json`, err)
	}
	if err = json.Unmarshal(params, &GSettings); err != nil {
		log.Fatalln(`Unmarshall`, err)
	}
	os.Chdir(dir)
	/*	if GPageTpl,err =template.ParseGlob(`template/*.tpl`); err!=nil {
			log.Fatalln( err )
		}
		if GPagePattern,err =template.ParseGlob(`pattern/*.tpl`); err!=nil {
			log.Fatalln( err )
		}*/
	log.Println("Start")
	//	go Send()

	http.HandleFunc(`/`, testnetHandler)
	http.Handle("/static/", http.FileServer(&assetfs.AssetFS{Asset: FileAsset, AssetDir: static.AssetDir, Prefix: ""}))

	http.ListenAndServe(fmt.Sprintf(":%d", GSettings.Port), nil)
	log.Println("Finish")
}
