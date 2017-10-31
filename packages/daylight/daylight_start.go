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
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/apiv2"
	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/daemons"
	"github.com/AplaProject/go-apla/packages/daylight/daemonsctl"
	"github.com/AplaProject/go-apla/packages/exchangeapi"
	"github.com/AplaProject/go-apla/packages/language"
	logtools "github.com/AplaProject/go-apla/packages/log"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/schema"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/static"
	"github.com/AplaProject/go-apla/packages/utils"

	"github.com/go-bindata-assetfs"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/thrust"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// FileAsset returns the body of the file
func FileAsset(name string) ([]byte, error) {
	if name := strings.Replace(name, "\\", "/", -1); name == `static/img/logo.`+utils.LogoExt {
		logofile := *utils.Dir + `/logo.` + utils.LogoExt
		if fi, err := os.Stat(logofile); err == nil && fi.Size() > 0 {
			return ioutil.ReadFile(logofile)
		} else if err != nil {
			log.WithFields(log.Fields{"path": logofile, "error": err, "type": consts.IOError}).Error("Reading logo file")
		}
	}
	return static.Asset(name)
}

func readConfig() {
	// read the config.ini
	config.Read()
	if *utils.TCPHost == "" {
		*utils.TCPHost = config.ConfigIni["tcp_host"]
	}
	if *utils.FirstBlockDir == "" {
		*utils.FirstBlockDir = config.ConfigIni["first_block_dir"]
	}
	if *utils.ListenHTTPPort == "" {
		*utils.ListenHTTPPort = config.ConfigIni["http_port"]
	}
	if *utils.Dir == "" {
		*utils.Dir = config.ConfigIni["dir"]
	}
	country, err := strconv.ParseInt(config.ConfigIni["one_country"], 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"value": config.ConfigIni["one_country"], "type": consts.ConvertionError}).Error("parsing to int")
	}
	utils.OneCountry = country
	utils.PrivCountry = config.ConfigIni["priv_country"] == `1` || config.ConfigIni["priv_country"] == `true`
	if len(config.ConfigIni["lang"]) > 0 {
		language.LangList = strings.Split(config.ConfigIni["lang"], `,`)
	}
}

func killOld() {
	pidPath := *utils.Dir + "/daylight.pid"
	if _, err := os.Stat(pidPath); err == nil {
		data, err := ioutil.ReadFile(pidPath)
		if err != nil {
			log.WithFields(log.Fields{"path": pidPath, "error": err, "type": consts.IOError}).Error("reading pid file")
		}
		var pidMap map[string]string
		err = json.Unmarshal(data, &pidMap)
		if err != nil {
			log.WithFields(log.Fields{"data": data, "error": err, "type": consts.JSONUnmarshallError}).Error("unmarshalling pid map")
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
	toMarshal := map[string]string{"pid": converter.IntToStr(pid), "version": consts.VERSION}
	PidAndVer, err := json.Marshal(toMarshal)
	if err != nil {
		log.WithFields(log.Fields{"pid": pid, "data": toMarshal, "error": err, "type": consts.JSONMarshallError}).Error("marshalling json")
		return err
	}
	return ioutil.WriteFile(*utils.Dir+"/daylight.pid", PidAndVer, 0644)
}

func delPidFile() {
	os.Remove(filepath.Join(*utils.Dir, "daylight.pid"))
}

func rollbackToBlock(blockID int64) error {
	if err := smart.LoadContracts(nil); err != nil {
		log.Errorf(`Load Contracts: %s`, err)
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
	startData := map[string]int64{"install": 1, "config": 1, "queue_tx": 99999, "log_transactions": 1, "transactions_status": 99999, "block_chain": 1, "info_block": 1, "dlt_wallets": 1, "confirmations": 9999999, "full_nodes": 1, "system_parameters": 4, "my_node_keys": 99999, "transactions": 999999}
	for _, table := range allTable {
		query := "SELECT COUNT(*) FROM " + converter.EscapeName(table)
		count, err := model.Single(query).Int64()
		if err != nil {
			log.WithFields(log.Fields{"error": err, "query": query, "type": consts.DBError}).Error("Error querying DB")
			return err
		}
		if count > 0 && count > startData[table] {
			log.WithFields(log.Fields{"count": count, "start_data": startData[table], "table": table}).Warn("record count in table is larger then start")
		} else {
			log.WithFields(log.Fields{"count": count, "start_data": startData[table], "table": table}).Info("record count in table is ok")
		}
	}
	return nil
}

func processOldFile(oldFileName string) error {
	err := utils.CopyFileContents(os.Args[0], oldFileName)
	if err != nil {
		return err
	}
	schema.Migration()

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
	setRoute(route, `/exchangeapi/:name`, exchangeapi.API, `GET`, `POST`)
	setRoute(route, `/monitoring`, daemons.Monitoring, `GET`)
	apiv2.Route(route)
	route.Handler(`GET`, `/static/*filepath`, http.FileServer(&assetfs.AssetFS{Asset: FileAsset, AssetDir: static.AssetDir, Prefix: ""}))
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
func Start(dir string, thrustWindowLoder *window.Window) {
	var err error

	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{"panic": r, "type": consts.PanicRecoveredError}).Error("recovered panic")
			panic(r)
		}
	}()

	Exit := func(code int) {
		if thrustWindowLoder != nil {
			thrustWindowLoder.Close()
		}
		model.GormClose()
		delPidFile()
		os.Exit(code)
	}

	if dir != "" {
		*utils.Dir = dir
	}

	readConfig()

	err = initLogs()
	if err != nil {
		Exit(1)
	}

	if len(config.ConfigIni["db_type"]) > 0 {
		// The installation process is already finished (where user has specified DB and where wallet has been restarted)
		err = model.GormInit(config.ConfigIni["db_user"], config.ConfigIni["db_password"], config.ConfigIni["db_name"])
		if err != nil {
			log.WithFields(log.Fields{"db_user": config.ConfigIni["db_user"], "db_password": config.ConfigIni["db_password"],
				"db_name": config.ConfigIni["db_name"], "type": consts.DBError}).Error("can't init gorm")
			Exit(1)
		}
	}

	// create first block
	if *utils.GenerateFirstBlock == 1 {
		log.Infof("generate first block")
		parser.FirstBlock()
		os.Exit(0)
	}

	log.WithFields(log.Fields{"work_dir": *utils.Dir, "version": consts.VERSION}).Info("started with")
	exchangeapi.InitAPI()
	log.Info("Initialized exchange API")

	// kill previously run apla
	if !utils.Mobile() {
		killOld()
	}

	// TODO: ??
	if fi, err := os.Stat(*utils.Dir + `/logo.png`); err == nil && fi.Size() > 0 {
		utils.LogoExt = `png`
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
			Exit(1)
		}
		defer delPidFile()
	}

	// database rollback to the specified block
	if *utils.RollbackToBlockID > 0 {
		log.WithFields(log.Fields{"block_id": *utils.RollbackToBlockID}).Info("Rollbacking to block ID")
		rollbackToBlock(*utils.RollbackToBlockID)
		log.WithFields(log.Fields{"block_id": *utils.RollbackToBlockID}).Info("Rollback is ok")
		Exit(0)
	}

	dir = *utils.Dir + "/public"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			log.WithFields(log.Fields{"path": dir, "error": err, "type": consts.IOError}).Error("Making dir")
			Exit(1)
		}
	}

	BrowserHTTPHost, _, ListenHTTPHost := GetHTTPHost()
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
		BrowserHTTPHost = initRoutes(ListenHTTPHost, BrowserHTTPHost)

		if *utils.Console == 0 && !utils.Mobile() {
			log.Info("starting browser")
			time.Sleep(time.Second)
			if thrustWindowLoder != nil {
				thrustWindowLoder.Close()
				thrustWindow := thrust.NewWindow(thrust.WindowOptions{
					RootUrl: BrowserHTTPHost,
					Size:    commands.SizeHW{Width: 1024, Height: 700},
				})
				if *utils.DevTools != 0 {
					thrustWindow.OpenDevtools()
				}
				thrustWindow.HandleEvent("*", func(cr commands.EventResult) {
					log.WithFields(log.Fields{"event": cr}).Debug("handle event")
				})
				thrustWindow.HandleRemote(func(er commands.EventResult, this *window.Window) {
					if len(er.Message.Payload) > 7 && er.Message.Payload[:7] == `mailto:` && runtime.GOOS == `windows` {
						utils.ShellExecute(er.Message.Payload)
					} else if len(er.Message.Payload) > 7 && er.Message.Payload[:2] == `[{` {
						ioutil.WriteFile(filepath.Join(*utils.Dir, `accounts.txt`), []byte(er.Message.Payload), 0644)
					} else if er.Message.Payload == `ACCOUNTS` {
						accounts, _ := ioutil.ReadFile(filepath.Join(*utils.Dir, `accounts.txt`))
						this.SendRemoteMessage(string(accounts))
					} else {
						openBrowser(er.Message.Payload)
					}
					// Keep in mind once we have the message, lets say its json of some new type we made,
					// We can unmarshal it to that type.
					// Same goes for the other way around.
				})
				thrustWindow.Show()
				thrustWindow.Focus()
			} else {
				//				openBrowser(BrowserHTTPHost)
			}
		}
	}()

	// waits for new records in chat, then waits for connect
	// (they are entered from the 'connections' daemon and from those who connected to the node by their own)
	// go utils.ChatOutput(utils.ChatNewTx)

	select {}
}
