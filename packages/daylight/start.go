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

package daylight

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AplaProject/go-apla/packages/api"
	conf "github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/daemons"
	"github.com/AplaProject/go-apla/packages/daylight/daemonsctl"
	logtools "github.com/AplaProject/go-apla/packages/log"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/statsd"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// !!! remove
// func readConfig() {
// 	// read the config.ini
// 	config.Read()
// 	if *utils.TCPHost == "" {
// 		*utils.TCPHost = config.ConfigIni["tcp_host"]
// 	}
// 	if *utils.FirstBlockDir == "" {
// 		*utils.FirstBlockDir = config.ConfigIni["first_block_dir"]
// 	}
// 	if *utils.ListenHTTPPort == "" {
// 		*utils.ListenHTTPPort = config.ConfigIni["http_port"]
// 	}
// 	if *utils.Dir == "" {
// 		*utils.Dir = config.ConfigIni["dir"]
// 	}
// 	utils.OneCountry = converter.StrToInt64(config.ConfigIni["one_country"])
// 	utils.PrivCountry = config.ConfigIni["priv_country"] == `1` || config.ConfigIni["priv_country"] == `true`
// 	if len(config.ConfigIni["lang"]) > 0 {
// 		language.LangList = strings.Split(config.ConfigIni["lang"], `,`)
// 	}
// }

func initStatsd() {
	// host := "127.0.0.1"
	// port := 8125
	// var name = "apla"
	// if config.ConfigIni["stastd_host"] != "" {
	// 	host = config.ConfigIni["statsd_host"]
	// }
	// if config.ConfigIni["stastd_port"] != "" {
	// 	port = converter.StrToInt(config.ConfigIni["statsd_port"])
	// }
	// if config.ConfigIni["statsd_client_name"] != "" {
	// 	name = config.ConfigIni["statsd_client_name"]
	// }
	cfg := conf.Config.StatsD
	if err := statsd.Init(cfg.Host, converter.StrToInt(cfg.Port), cfg.Name); err != nil {
		log.WithFields(log.Fields{"type": consts.StatsdError, "error": err}).Fatal("cannot initialize statsd")
	}
}

func killOld() {
	pidPath := *utils.Dir + "/daylight.pid"
	if _, err := os.Stat(pidPath); err == nil {
		dat, err := ioutil.ReadFile(pidPath)
		if err != nil {
			log.WithFields(log.Fields{"path": pidPath, "error": err, "type": consts.IOError}).Error("reading pid file")
		}
		var pidMap map[string]string
		err = json.Unmarshal(dat, &pidMap)
		if err != nil {
			log.WithFields(log.Fields{"data": dat, "error": err, "type": consts.JSONUnmarshallError}).Error("unmarshalling pid map")
		}
		log.WithFields(log.Fields{"path": *utils.Dir + pidMap["pid"]}).Debug("old pid path")

		KillPid(pidMap["pid"])
		if fmt.Sprintf("%s", err) != "null" {
			// give 15 sec to end the previous process
			for i := 0; i < 15; i++ {
				if _, err := os.Stat(*utils.Dir + "/daylight.pid"); err == nil {
					time.Sleep(time.Second)
				} else { // if there is no daylight.pid, so it is finished
					break
				}
			}
		}
	}
}

func initLogs() error {
	var err error

	if config.ConfigIni["log_output"] != "file" {
		log.SetOutput(os.Stdout)
	} else {
		fileName := *utils.Dir + "/dclog.txt"
		openMode := os.O_APPEND
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			openMode = os.O_CREATE
		}

		f, err := os.OpenFile(fileName, os.O_WRONLY|openMode, 0755)
		if err != nil {
			fmt.Println("Can't open log file ", fileName)
			return err
		}
		log.SetOutput(f)
	}

	if level, ok := config.ConfigIni["log_level"]; ok {
		switch level {
		case "Debug":
			log.SetLevel(log.DebugLevel)
		case "Info":
			log.SetLevel(log.InfoLevel)
		case "Warn":
			log.SetLevel(log.WarnLevel)
		case "Error":
			log.SetLevel(log.ErrorLevel)
		}
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.AddHook(logtools.ContextHook{})
	return err
}

func savePid() error {
	pid := os.Getpid()
	PidAndVer, err := json.Marshal(map[string]string{"pid": converter.IntToStr(pid), "version": consts.VERSION})
	if err != nil {
		log.WithFields(log.Fields{"pid": pid, "error": err, "type": consts.JSONMarshallError}).Error("marshalling pid to json")
		return err
	}
	return ioutil.WriteFile(*utils.Dir+"/daylight.pid", PidAndVer, 0644)
}

func delPidFile() {
	os.Remove(filepath.Join(*utils.Dir, "daylight.pid"))
}

func rollbackToBlock(blockID int64) error {
	if err := smart.LoadContracts(nil); err != nil {
		return err
	}
	parser := new(parser.Parser)
	err := parser.RollbackToBlockID(*utils.RollbackToBlockID)
	if err != nil {
		return err
	}

	// we recieve the statistics of all tables
	allTable, err := model.GetAllTables()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("error getting all tables")
		return err
	}

	// block id = 1, is a special case for full rollback
	if blockID != 1 {
		return nil
	}

	// check blocks related tables
	startData := map[string]int64{"1_menu": 1, "1_pages": 1, "1_contracts": 26, "1_parameters": 11, "1_keys": 1, "1_tables": 8, "stop_daemons": 1, "queue_blocks": 9999999, "system_tables": 1, "system_parameters": 27, "system_states": 1, "install": 1, "config": 1, "queue_tx": 9999999, "log_transactions": 1, "transactions_status": 9999999, "block_chain": 1, "info_block": 1, "confirmations": 9999999, "transactions": 9999999}
	warn := 0
	for _, table := range allTable {
		count, err := model.GetRecordsCount(table)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting record count")
			return err
		}
		if count > 0 && count > startData[table] {
			log.WithFields(log.Fields{"count": count, "start_data": startData[table], "table": table}).Warn("record count in table is larger then start")
			warn++
		} else {
			log.WithFields(log.Fields{"count": count, "start_data": startData[table], "table": table}).Info("record count in table is ok")
		}
	}
	if warn == 0 {
		ioutil.WriteFile(*utils.Dir+"rollback_result", []byte("1"), 0644)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.WritingFile}).Error("write to the rollback_result")
			return err
		}
	}
	return nil
}

func processOldFile(oldFileName string) error {

	err := utils.CopyFileContents(os.Args[0], oldFileName)
	if err != nil {
		log.Errorf("can't copy from %s %v", oldFileName, utils.ErrInfo(err))
		return err
	}

	err = exec.Command(*utils.OldFileName, "-dir", *utils.Dir).Start()
	if err != nil {
		log.WithFields(log.Fields{"cmd": *utils.OldFileName + " -dir " + *utils.Dir, "error": err, "type": consts.CommandExecutionError}).Error("executing command")
		return err
	}
	return nil
}

func setRoute(route *httprouter.Router, path string, handle func(http.ResponseWriter, *http.Request), methods ...string) {
	for _, method := range methods {
		route.HandlerFunc(method, path, handle)
	}
}
func initRoutes(listenHost, browserHost string) string {
	route := httprouter.New()
	setRoute(route, `/monitoring`, daemons.Monitoring, `GET`)
	api.Route(route)
	route.Handler(`GET`, `/.well-known/*filepath`, http.FileServer(http.Dir(*utils.TLS)))
	if len(*utils.TLS) > 0 {
		go http.ListenAndServeTLS(":443", *utils.TLS+`/fullchain.pem`, *utils.TLS+`/privkey.pem`, route)
	}

	httpListener(listenHost, &browserHost, route)
	// for ipv6 server
	httpListenerV6(route)
	return browserHost
}

// Start starts the main code of the program
func Start() {

	var err error

	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{"panic": r, "type": consts.PanicRecoveredError}).Error("recovered panic")
			panic(r)
		}
	}()

	Exit := func(code int) {
		model.GormClose()
		delPidFile()
		os.Exit(code)
		statsd.Close()
	}

	// // // // // // // // // // // // //

	fmt.Println("Start.") // !!!

	conf.ParseFlags()

	// parse flags

	// if initConfig

	// load toml config
	// apply flags

	if err := conf.LoadConfig(); err != nil {
		log.Error("loadConfig:", err)
		return
	}

	if err := conf.SaveConfig(); err != nil {
		log.Error("saveConfig:", err)
		return
	}
	return // !!!

	//	readConfig()

	// // // // // // // // // // // // //

	if len(config.ConfigIni["db_type"]) > 0 {
		// The installation process is already finished (where user has specified DB and where wallet has been restarted)
		err = model.GormInit(
			config.ConfigIni["db_host"], config.ConfigIni["db_port"],
			config.ConfigIni["db_user"], config.ConfigIni["db_password"], config.ConfigIni["db_name"])
		if err != nil {
			log.WithFields(log.Fields{"db_user": config.ConfigIni["db_user"], "db_password": config.ConfigIni["db_password"],
				"db_name": config.ConfigIni["db_name"], "type": consts.DBError}).Error("can't init gorm")
			Exit(1)
		}
	}

	// create first block
	if *utils.GenerateFirstBlock == 1 {
		log.Info("Generating first block")
		parser.FirstBlock()
		os.Exit(0)
	}

	log.WithFields(log.Fields{"work_dir": *utils.Dir, "version": consts.VERSION}).Info("started with")

	// kill previously run apla
	if !utils.Mobile() {
		killOld()
	}

	// TODO: ??
	if fi, err := os.Stat(*utils.Dir + `/logo.png`); err == nil && fi.Size() > 0 {
		utils.LogoExt = `png`
	}

	initStatsd()
	err = initLogs()
	if err != nil {
		fmt.Printf("logs init failed: %v\n", utils.ErrInfo(err))
		Exit(1)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	// if there is OldFileName, so act on behalf dc.tmp and we have to restart on behalf the normal name
	if *utils.OldFileName != "" {
		processOldFile(*utils.OldFileName)
		Exit(1)
	}

	// save the current pid and version
	if !utils.Mobile() {
		if err := savePid(); err != nil {
			log.Errorf("can't create pid: %s", err)
			Exit(1)
		}
		defer delPidFile()
	}

	// database rollback to the specified block
	if *utils.RollbackToBlockID > 0 {
		err = syspar.SysUpdate()
		if err != nil {
			log.WithError(err).Error("can't read system parameters")
		}
		log.WithFields(log.Fields{"block_id": *utils.RollbackToBlockID}).Info("Rollbacking to block ID")
		err := rollbackToBlock(*utils.RollbackToBlockID)
		log.WithFields(log.Fields{"block_id": *utils.RollbackToBlockID}).Info("Rollback is ok")
		if err != nil {
			log.WithError(err).Error("Rollback error")
		} else {
			log.Info("Rollback is OK")
		}
		Exit(0)
	}

	if _, err := os.Stat(*utils.Dir + "/public"); os.IsNotExist(err) {
		err = os.Mkdir(*utils.Dir+"/public", 0755)
		if err != nil {
			log.WithFields(log.Fields{"path": *utils.Dir, "error": err, "type": consts.IOError}).Error("Making dir")
			Exit(1)
		}
	}

	BrowserHTTPHost, ListenHTTPHost := GetHTTPHost()
	if model.DBConn != nil {
		// The installation process is already finished (where user has specified DB and where wallet has been restarted)
		err := daemonsctl.RunAllDaemons()
		log.Info("Daemons started")
		if err != nil {
			os.Exit(1)
		}
	}

	daemons.WaitForSignals()

	go func() {
		time.Sleep(time.Second)
		BrowserHTTPHost = initRoutes(ListenHTTPHost, BrowserHTTPHost) // !!! BrowserHTTPHost unused

		if *utils.Console == 0 && !utils.Mobile() {
			log.Info("starting browser")
			time.Sleep(time.Second)
		}
	}()

	// waits for new records in chat, then waits for connect
	// (they are entered from the 'connections' daemon and from those who connected to the node by their own)
	// go utils.ChatOutput(utils.ChatNewTx)

	select {}
}
