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
	"sync"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/astaxie/beego/config"
	"github.com/go-bindata-assetfs"
)

const (
	// XForwardedFor is a header parameter
	XForwardedFor = "X-Forwarded-For"
	// XRealIP is a header parameter
	XRealIP = "X-Real-IP"
)

// Settings contains options of the program
type Settings struct {
	Port uint32
	Path string
	Node string
}

// TestnetPage is a structure for the template of testnet handler
type TestnetPage struct {
	List   []*TxInfo
	Latest int64
	Host   string
	Node   string
}

// RegisterPage is a structure for the template of register handler
type RegisterPage struct {
	Available int64
	Country   string
	State     int64
	Node      string
	Message   string
}

// NewStateResult is a structure for the answer of newState handler
type NewStateResult struct {
	Private string `json:"private"`
	Wallet  string `json:"wallet"`
	Result  int64  `json:"result"`
	Error   string `json:"error"`
}

// NewRegisterResult is a structure for the answer of newregister handler
type NewRegisterResult struct {
	Private string `json:"private"`
	//	Wallet  string `json:"wallet"`
	//	Result int64  `json:"result"`
	Error string `json:"error"`
}

var (
	gSettings Settings
	regMutex  = &sync.Mutex{}
)

// FileAsset returns the content of the file
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

func escape(name string) string {
	out := make([]byte, 0, len(name))
	skip := `<>"'`
	for _, ch := range []byte(name) {
		if strings.IndexByte(skip, ch) < 0 {
			out = append(out, ch)
		}
	}
	return string(out)
}

func newstateHandler(w http.ResponseWriter, r *http.Request) {
	var result NewStateResult

	errFunc := func(msg string) {
		w.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, lib.EscapeForJSON(msg))))
	}

	r.ParseForm()
	email := strings.TrimSpace(r.FormValue(`email`))
	currency := escape(strings.TrimSpace(r.FormValue(`currency`)))
	country := escape(strings.TrimSpace(r.FormValue(`country`)))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if len(email) == 0 || !utils.ValidateEmail(email) {
		errFunc(`Email is not valid`)
		return
	}
	if len(country) == 0 {
		errFunc(`Country cannot be empty`)
		return
	}
	if len(currency) == 0 {
		errFunc(`Currency cannot be empty`)
		return
	}
	if id, err := utils.DB.Single(`select id from global_states_list where state_name=?`, country).Int64(); err != nil {
		errFunc(err.Error())
		return
	} else if id > 0 {
		errFunc(fmt.Sprintf(`State %s already exists`, country))
		return
	}
	if id, err := utils.DB.Single(`select id from global_currencies_list where currency_code=?`, currency).Int64(); err != nil {
		errFunc(err.Error())
		return
	} else if id > 0 {
		errFunc(fmt.Sprintf(`Currency %s already exists`, currency))
		return
	}
	if exist, err := utils.DB.Single(`select id from testnet_emails where email=? and country = ? and currency=?`,
		email, country, currency).Int64(); err != nil {
		errFunc(err.Error())
		return
	} else if exist > 0 {
		errFunc(fmt.Sprintf(`The same request has been already sent`))
		return
	}
	id, err := utils.DB.ExecSqlGetLastInsertId(`insert into testnet_emails (email,country,currency) 
				values(?,?,?)`, `testnet_emails`, email, country, currency)
	if err != nil {
		result.Error = err.Error()
	} else {
		result.Result = utils.StrToInt64(id)
		resp, err := http.Get(strings.TrimRight(gSettings.Node, `/`) + `/ajax?json=ajax_new_state&testnet=` + id)
		if err != nil {
			errFunc(err.Error())
			return
		}
		defer resp.Body.Close()
		var answer []byte
		if answer, err = ioutil.ReadAll(resp.Body); err != nil {
			errFunc(err.Error())
			return
		}
		var answerJSON NewStateResult
		if err = json.Unmarshal(answer, &answerJSON); err != nil {
			errFunc(err.Error() + `; ` + string(answer))
			return
		}
		if answerJSON.Error != `success` {
			errFunc(answerJSON.Error)
			return
		}
		upd, err := utils.DB.OneRow(`select private, wallet from testnet_emails where id=?`, id).String()
		if err != nil {
			errFunc(err.Error())
			return
		}
		result.Private = upd[`private`]
		result.Wallet = lib.AddressToString(utils.StrToInt64(upd[`wallet`]))
	}

	if jsonData, err := json.Marshal(result); err == nil {
		w.Write(jsonData)
	} else {
		w.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
	}
}

func newregisterHandler(w http.ResponseWriter, r *http.Request) {
	var result NewRegisterResult

	errFunc := func(msg string) {
		w.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, lib.EscapeForJSON(msg))))
	}

	r.ParseForm()
	state := utils.StrToInt64(r.FormValue(`state`))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if state == 0 {
		errFunc(`State is not defined`)
		return
	}
	regMutex.Lock()
	defer regMutex.Unlock()

	key, err := utils.DB.OneRow(`select id, wallet, private from testnet_keys where state_id=? and status=0`, state).String()
	if err != nil {
		errFunc(err.Error())
		return
	}
	if len(key) == 0 || len(key[`private`]) == 0 {
		errFunc(`There are not available keys`)
		return
	}
	err = utils.DB.ExecSql(`update testnet_keys set status=1 where id=? and state_id=? and status=0 and wallet=?`,
		key[`id`], state, key[`wallet`])
	if err != nil {
		errFunc(err.Error())
		return
	}
	result.Private = key[`private`]
	if jsonData, err := json.Marshal(result); err == nil {
		w.Write(jsonData)
	} else {
		w.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
	}
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
	list := make([]*TxInfo, 0)
	start := txTop
	for start.ID > 0 && len(list) < 10 {
		list = append(list, start)
		start = start.prev
	}
	err = t.Execute(b, TestnetPage{Node: strings.TrimRight(gSettings.Node, `/`), List: list,
		Latest: txTop.ID, Host: r.Host})
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
	}
	w.Write(b.Bytes())
	return
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	var (
		message, country string
		available, state int64
		err              error
	)
	state = utils.StrToInt64(r.FormValue(`state`))
	if state == 0 {
		message = `State parameter is not defined`
	} else {
		country, err = utils.DB.Single(`select state_name from global_states_list where gstate_id=?`, state).String()
		if err != nil {
			message = err.Error()
		} else if len(country) == 0 {
			message = fmt.Sprintf(`State %d has not been found`, state)
		} else {
			available, err = utils.DB.Single(`select count(id) from testnet_keys where state_id=? and status=0`, state).Int64()
			if err != nil {
				message = err.Error()
			} else if available == 0 {
				message = `Unfortunately, there are not available private keys for registering. Try to open this link later.`
			}
		}
	}

	data, err := static.Asset("static/testnet_register.html")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
	}
	t, err := template.New("template").Funcs(funcMap).Parse(string(data))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
	}
	b := new(bytes.Buffer)
	err = t.Execute(b, RegisterPage{Node: strings.TrimRight(gSettings.Node, `/`), Available: available,
		Country: country, State: state, Message: message})
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
	if err = json.Unmarshal(params, &gSettings); err != nil {
		log.Fatalln(`Unmarshall`, err)
	}
	os.Chdir(dir)
	*utils.Dir = gSettings.Path
	configIni := make(map[string]string)
	fullConfigIni, err := config.NewConfig("ini", gSettings.Path+`/config.ini`)
	if err != nil {
		log.Fatalln(`Config`, err)
	} else {
		configIni, err = fullConfigIni.GetSection("default")
	}
	if utils.DB, err = utils.NewDbConnect(configIni); err != nil {
		log.Fatalln(`Utils connection`, err)
	}
	list, err := utils.DB.GetAllTables()
	if err != nil || len(list) == 0 {
		log.Fatalln(`GetAllTables`, err)
	}
	if !utils.InSliceString(`testnet_emails`, list) {
		if err = utils.DB.ExecSql(`CREATE SEQUENCE testnet_emails_id_seq START WITH 1;
CREATE TABLE "testnet_emails" (
"id" integer NOT NULL DEFAULT nextval('testnet_emails_id_seq'),
"email" varchar(128) NOT NULL DEFAULT '',
"country" varchar(128) NOT NULL DEFAULT '',
"currency" varchar(32) NOT NULL DEFAULT '',
"private" varchar(64) NOT NULL DEFAULT '',
"wallet" bigint NOT NULL DEFAULT '0',
"status" integer NOT NULL DEFAULT '0',
"code" integer NOT NULL DEFAULT '0',
"validate" integer NOT NULL DEFAULT '0'
);
ALTER SEQUENCE testnet_emails_id_seq owned by testnet_emails.id;
ALTER TABLE ONLY "testnet_emails" ADD CONSTRAINT testnet_emails_pkey PRIMARY KEY (id);
CREATE INDEX testnet_index_email ON "testnet_emails" (email);`); err != nil {
			log.Fatalln(err)
		}
	}
	/*	if !utils.InSliceString(`global_currencies_list`, list) {
				if err = utils.DB.ExecSql(`CREATE SEQUENCE global_currencies_list_id_seq START WITH 1;
		CREATE TABLE "global_currencies_list" (
		"id" integer NOT NULL DEFAULT nextval('global_currencies_list_id_seq'),
		"currency_code" varchar(32) NOT NULL DEFAULT '',
		"settings_table" varchar(128) NOT NULL DEFAULT '',
		"rb_id" int NOT NULL DEFAULT '0'
		);
		ALTER SEQUENCE global_currencies_list_id_seq owned by global_currencies_list.id;
		ALTER TABLE ONLY "global_currencies_list" ADD CONSTRAINT global_currencies_list_pkey PRIMARY KEY (id);
		CREATE INDEX global_currencies_index_code ON "global_currencies_list" (currency_code);`); err != nil {
					log.Fatalln(err)
				}
				if states, err := utils.DB.GetAll(`select * from system_states order by id`, -1); err != nil {
					log.Fatalln(err)
				} else {
					for _, item := range states {
						table := item[`id`] + `_state_parameters`
						if code, err := utils.DB.Single(`select value from "` + table + `" where name='currency_name'`).String(); err != nil {
							log.Fatalln(err)
						} else {
							if err = utils.DB.ExecSql(`insert into global_currencies_list (currency_code, settings_table)
							    values(?,?)`, code, table); err != nil {
								log.Fatalln(err)
							}
						}
					}
				}
			}
			if !utils.InSliceString(`global_states_list`, list) {
				if err = utils.DB.ExecSql(`CREATE SEQUENCE global_states_list_id_seq START WITH 1;
		CREATE TABLE "global_states_list" (
		"id" integer NOT NULL DEFAULT nextval('global_states_list_id_seq'),
		"state_id" bigint NOT NULL DEFAULT '0',
		"state_name" varchar(128) NOT NULL DEFAULT '',
		"rb_id" int NOT NULL DEFAULT '0'
		);
		ALTER SEQUENCE global_states_list_id_seq owned by global_states_list.id;
		ALTER TABLE ONLY "global_states_list" ADD CONSTRAINT global_states_list_pkey PRIMARY KEY (id);
		CREATE INDEX global_states_index_name ON "global_states_list" (state_name);`); err != nil {
					log.Fatalln(err)
				}
				if states, err := utils.DB.GetAll(`select * from system_states order by id`, -1); err != nil {
					log.Fatalln(err)
				} else {
					for _, item := range states {
						table := item[`id`] + `_state_parameters`
						if state, err := utils.DB.Single(`select value from "` + table + `" where name='state_name'`).String(); err != nil {
							log.Fatalln(err)
						} else {
							if err = utils.DB.ExecSql(`insert into global_states_list (state_id, state_name)
							    values(?,?)`, item[`id`], state); err != nil {
								log.Fatalln(err)
							}
						}
					}
				}

			}*/
	if !utils.InSliceString(`testnet_keys`, list) {
		if err = utils.DB.ExecSql(`CREATE TABLE "testnet_keys" (
		"id" bigint NOT NULL DEFAULT '0',
		"state_id" integer NOT NULL DEFAULT '0',
		"private" varchar(64) NOT NULL DEFAULT '',
		"wallet" bigint NOT NULL DEFAULT '0',
		"status" integer NOT NULL DEFAULT '0'
		);
		CREATE INDEX testnet_index_keys ON "testnet_keys" (id,state_id,status);`); err != nil {
			log.Fatalln(err)
		}
	}
	log.Println("Start")
	go GetTx()

	http.HandleFunc(`/`, testnetHandler)
	http.HandleFunc(`/newstate`, newstateHandler)
	http.HandleFunc(`/newregister`, newregisterHandler)
	http.HandleFunc(`/register`, registerHandler)
	http.HandleFunc(`/wschain`, WsBlockchain)
	http.Handle("/static/", http.FileServer(&assetfs.AssetFS{Asset: FileAsset, AssetDir: static.AssetDir, Prefix: ""}))

	err = http.ListenAndServe(fmt.Sprintf(":%d", gSettings.Port), nil)
	if err != nil {
		log.Fatalln(`Listen`, err)
	}
	log.Println("Finish")
}
