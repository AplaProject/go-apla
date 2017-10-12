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
	//	_ "image/png"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/apiv2"
	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/daemons"
	"github.com/AplaProject/go-apla/packages/exchangeapi"
	"github.com/AplaProject/go-apla/packages/language"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/schema"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/static"
	"github.com/AplaProject/go-apla/packages/stopdaemons"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/go-bindata-assetfs"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/thrust"
	"github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
)

// FileAsset returns the body of the file
func FileAsset(name string) ([]byte, error) {

	if name := strings.Replace(name, "\\", "/", -1); name == `static/img/logo.`+utils.LogoExt {
		logofile := *utils.Dir + `/logo.` + utils.LogoExt
		if fi, err := os.Stat(logofile); err == nil && fi.Size() > 0 {
			return ioutil.ReadFile(logofile)
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
	utils.OneCountry = converter.StrToInt64(config.ConfigIni["one_country"])
	utils.PrivCountry = config.ConfigIni["priv_country"] == `1` || config.ConfigIni["priv_country"] == `true`
	if len(config.ConfigIni["lang"]) > 0 {
		language.LangList = strings.Split(config.ConfigIni["lang"], `,`)
	}
}

func killOld() {
	if _, err := os.Stat(*utils.Dir + "/daylight.pid"); err == nil {
		dat, err := ioutil.ReadFile(*utils.Dir + "/daylight.pid")
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
		var pidMap map[string]string
		err = json.Unmarshal(dat, &pidMap)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
		fmt.Println("old PID ("+*utils.Dir+"/daylight.pid"+"):", pidMap["pid"])

		err = KillPid(pidMap["pid"])
		if nil != err {
			fmt.Println(err)
			log.Error("KillPid %v", utils.ErrInfo(err))
		}
		if fmt.Sprintf("%s", err) != "null" {
			fmt.Println(fmt.Sprintf("%s", err))
			// give 15 sec to end the previous process
			for i := 0; i < 15; i++ {
				log.Debug("waiting killer %d", i)
				if _, err := os.Stat(*utils.Dir + "/daylight.pid"); err == nil {
					fmt.Println("waiting killer")
					time.Sleep(time.Second)
				} else { // if there is no daylight.pid, so it is finished
					break
				}
			}
		}
	}
}

func initLogs() error {
	var backend *logging.LogBackend

	if config.ConfigIni["log_output"] == "console" {
		backend = logging.NewLogBackend(os.Stderr, "", 0)
	} else {
		f, err := os.OpenFile(*utils.Dir+"/dclog.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
		if err != nil {
			return err
		}
		backend = logging.NewLogBackend(f, "", 0)
	}

	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)

	if *utils.LogLevel == "" {
		if level, ok := config.ConfigIni["log_level"]; ok {
			*utils.LogLevel = level
		} else {
			*utils.LogLevel = "INFO"
		}
	}

	logLevel, err := logging.LogLevel(*utils.LogLevel)
	if err != nil {
		log.Error("bad log level - %s: %v", *utils.LogLevel, utils.ErrInfo(err))
		return err
	}

	log.Infof("set logLevel: %v", logLevel)
	backendLeveled.SetLevel(logLevel, "")
	logging.SetBackend(backendLeveled)
	return nil
}

func savePid() error {
	pid := os.Getpid()
	PidAndVer, err := json.Marshal(map[string]string{"pid": converter.IntToStr(pid), "version": consts.VERSION})
	if err != nil {
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
		log.Errorf("rollback return error: %s", err)
		return err
	}

	// we recieve the statistics of all tables
	allTable, err := model.GetAllTables()
	if err != nil {
		log.Errorf("get all tables failed: %s", err)
		return err
	}

	// block id = 1, is a special case for full rollback
	if blockID != 1 {
		return nil
	}

	// check blocks related tables
	startData := map[string]int64{"install": 1, "config": 1, "queue_tx": 99999, "log_transactions": 1, "transactions_status": 99999, "block_chain": 1, "info_block": 1, "dlt_wallets": 1, "confirmations": 9999999, "full_nodes": 1, "system_parameters": 4, "my_node_keys": 99999, "transactions": 999999}
	for _, table := range allTable {
		count, err := model.Single(`SELECT count(*) FROM ` + converter.EscapeName(table)).Int64()
		if err != nil {
			log.Errorf("select from table %s failed: %s", table, err)
			return err
		}
		if count > 0 && count > startData[table] {
			fmt.Println(">>ALERT<<", table, count)
		} else {
			fmt.Println(table, "ok")
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
	schema.Migration()

	err = exec.Command(*utils.OldFileName, "-dir", *utils.Dir).Start()
	if err != nil {
		log.Errorf("exec command error: %v", utils.ErrInfo(err))
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
	IosLog("start")

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
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

	if len(config.ConfigIni["db_type"]) > 0 {
		// The installation process is already finished (where user has specified DB and where wallet has been restarted)
		err = model.GormInit(config.ConfigIni["db_user"], config.ConfigIni["db_password"], config.ConfigIni["db_name"])
		if err != nil {
			log.Errorf("gorm init error: %s", err)
			Exit(1)
		}

		err = syspar.SysUpdate()
		if err != nil {
			log.Error("can't read system parameters: %s", utils.ErrInfo(err))
			Exit(1)
		}
	}

	// create first block
	if *utils.GenerateFirstBlock == 1 {
		log.Infof("generate first block")
		parser.FirstBlock()
		os.Exit(0)

	}

	fmt.Printf("work dir = %s\ndcVersion=%s\n", *utils.Dir, consts.VERSION)

	exchangeapi.InitAPI()

	// kill previously run apla
	if !utils.Mobile() {
		killOld()
	}

	// TODO: ??
	if fi, err := os.Stat(*utils.Dir + `/logo.png`); err == nil && fi.Size() > 0 {
		utils.LogoExt = `png`
	}

	err = initLogs()
	if err != nil {
		log.Error("logs init failed: %v", utils.ErrInfo(err))
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
		err := rollbackToBlock(*utils.RollbackToBlockID)
		if err != nil {
			fmt.Printf("rollback error: %s\n", err)
		} else {
			fmt.Printf("OK\n")
		}
		Exit(0)
	}

	log.Debug("public")
	IosLog("public")
	if _, err := os.Stat(*utils.Dir + "/public"); os.IsNotExist(err) {
		err = os.Mkdir(*utils.Dir+"/public", 0755)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			Exit(1)
		}
	}

	BrowserHTTPHost, _, ListenHTTPHost := GetHTTPHost()
	fmt.Printf("BrowserHTTPHost: %v, ListenHTTPHost: %v\n", BrowserHTTPHost, ListenHTTPHost)

	if model.DBConn != nil {
		// The installation process is already finished (where user has specified DB and where wallet has been restarted)
		log.Info("start daemons")
		daemons.StartDaemons()
		log.Debugf("daemon started")

		daemonsTable := make(map[string]string)
		go func() {
			for {
				daemonNameAndTime := <-daemons.MonitorDaemonCh
				daemonsTable[daemonNameAndTime[0]] = daemonNameAndTime[1]
				if time.Now().Unix()%10 == 0 {
					log.Debug("daemonsTable: %v\n", daemonsTable)
				}
			}
		}()

		// signals for daemons to exit
		go stopdaemons.WaitStopTime()

		if err := smart.LoadContracts(nil); err != nil {
			log.Errorf("Load Contracts error: %s", err)
			Exit(1)
		}
		log.Debugf("all contracts loaded")

		tcpListener()
		log.Debugf("tcp listener started")
	}

	stopdaemons.WaintForSignals()

	go func() {
		time.Sleep(time.Second)
		BrowserHTTPHost = initRoutes(ListenHTTPHost, BrowserHTTPHost)

		if *utils.Console == 0 && !utils.Mobile() {
			log.Debugf("try to start browser")
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
					fmt.Println("HandleEvent", cr)
				})
				thrustWindow.HandleRemote(func(er commands.EventResult, this *window.Window) {
					//					fmt.Println("RemoteMessage Recieved:", er.Message.Payload)
					if len(er.Message.Payload) > 7 && er.Message.Payload[:7] == `mailto:` && runtime.GOOS == `windows` {
						utils.ShellExecute(er.Message.Payload)
					} else if len(er.Message.Payload) > 7 && er.Message.Payload[:2] == `[{` {
						ioutil.WriteFile(filepath.Join(*utils.Dir, `accounts.txt`), []byte(er.Message.Payload), 0644)
						//					} else if len(er.Message.Payload) >= 7 && er.Message.Payload[:7] == `USERID=` {
						// for Lite version - do nothing
					} else if er.Message.Payload == `ACCOUNTS` {
						accounts, _ := ioutil.ReadFile(filepath.Join(*utils.Dir, `accounts.txt`))
						this.SendRemoteMessage(string(accounts))
					} else {
						openBrowser(er.Message.Payload)
					}
					// Keep in mind once we have the message, lets say its json of some new type we made,
					// We can unmarshal it to that type.
					// Same goes for the other way around.
					//					this.SendRemoteMessage("boop")
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

	time.Sleep(time.Second * 3600 * 24 * 90)
	log.Infof("exit")
}
