// statserv
package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/DayLightProject/go-daylight/packages/stat"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io/ioutil"
	"log"
	"strings"
	"net"
	"net/http"
	"os"
	"path/filepath"
	//	"regexp"
	//	"net/url"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

type Settings struct {
	Port uint32 `json:"port"`
	Path string `json:"path"`
	Period uint32 `json:"period"`
}

var (
	GSettings Settings
	GDB       *utils.DCDB
)

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
	return ipval,remoteAddr
}

func historyBalance(userId int64, history *stat.HistoryBalance) error {
	data, err := GDB.GetAll(`select * from balance where user_id=? AND date( uptime ) < date('now') order by uptime desc`,
		7, userId)
	if err == nil {
		for _, idata := range data {
			var ib stat.InfoBalance
			if err = json.Unmarshal([]byte(idata[`data`]), &ib); err == nil {
				history.History = append(history.History, &ib)
			} else {
				return err
			}
		}
	}
	return err
}

func balanceHandler(w http.ResponseWriter, r *http.Request) {

	answer := stat.HistoryBalance{true, ``, make([]*stat.InfoBalance, 0)}

	userId := utils.StrToInt64(r.FormValue(`user_id`))
	if userId > 0 {
		ipval,_ := getIP( r )
		err := historyBalance(userId, &answer)
		if err!=nil {
			log.Println(err)
		}
		GDB.ExecSql(`insert into req_balance ( user_id, ip, uptime) values( ?, ?,  datetime('now'))`,
							userId, ipval)
	}
	ret, err := json.Marshal(answer)
	if err != nil {
		ret = []byte(`{"success": false,
					   "error":"Unknown error"}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(ret)
}

func statHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(`{"success": true}`))
}

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(`Dir`, err)
	}
	//	os.Chdir(dir)
	logfile, err := os.OpenFile(filepath.Join(dir, "stat.log"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(`Stat log`, err)
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
	if err = os.Chdir(GSettings.Path); err != nil {
		log.Fatalln(`Chdir`, err)
	}
	if GDB, err = utils.NewDbConnect(map[string]string{
		"db_name": "", "db_password": ``, `db_port`: ``,
		`db_user`: ``, `db_host`: ``, `db_type`: `sqlite`}); err != nil {
		log.Fatalln(`Connect`, err)
	}

	*utils.Dir = GSettings.Path
	configIni := make(map[string]string)
	configIni_, err := config.NewConfig("ini", `config.ini`)
	if err != nil {
		log.Fatalln(`Config`, err)
	} else {
		configIni, err = configIni_.GetSection("default")
	}
	if utils.DB, err = utils.NewDbConnect(configIni); err != nil {
		log.Fatalln(`Utils connect`, err)
	}

	var list []string
	if list, err = GDB.GetAllTables(); err == nil && len(list) == 0 {
		if err = GDB.ExecSql(`CREATE TABLE balance (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_id	INTEGER NOT NULL,
	data    TEXT NOT NULL,
	uptime	INTEGER NOT NULL
	)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX userid ON balance (user_id,uptime)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE TABLE req_balance (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_id	INTEGER NOT NULL,
	ip	INTEGER NOT NULL,
	uptime	INTEGER NOT NULL
	)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX req_userid ON req_balance (user_id)`); err != nil {
			log.Fatalln(err)
		}
	}
	os.Chdir(dir)
	go daemon()

	log.Println("Start")

	http.HandleFunc(`/`, statHandler)
	http.HandleFunc(`/balance`, balanceHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", GSettings.Port), nil)
	log.Println("Finish")
}
