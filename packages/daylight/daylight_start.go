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

	"github.com/EGaaS/go-egaas-mvp/packages/api"
	"github.com/EGaaS/go-egaas-mvp/packages/config"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/controllers"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/daemons"
	"github.com/EGaaS/go-egaas-mvp/packages/exchangeapi"
	"github.com/EGaaS/go-egaas-mvp/packages/language"
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/schema"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/stopdaemons"
	"github.com/EGaaS/go-egaas-mvp/packages/system"
	"github.com/EGaaS/go-egaas-mvp/packages/template"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	"github.com/go-bindata-assetfs"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/thrust"
	"github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
)

func setRoute(route *httprouter.Router, path string, handle func(http.ResponseWriter, *http.Request), methods ...string) {
	for _, method := range methods {
		route.HandlerFunc(method, path, handle)
	}
}

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
		os.Exit(code)
	}

	if dir != "" {
		fmt.Println("dir", dir)
		*utils.Dir = dir
	}

	utils.FirstBlock(true)

	IosLog("dir:" + dir)
	fmt.Println("utils.Dir", *utils.Dir)

	fmt.Println("dcVersion:", consts.VERSION)
	log.Debug("dcVersion: %v", consts.VERSION)

	exchangeapi.InitAPI()

	// читаем config.ini
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
		language.LangList = strings.Split(configIni["lang"], `,`)
	}
	/*	outfile, err := os.Create("./out.txt")
	    if err != nil {
	        panic(err)
	    }
	    defer outfile.Close()
		os.Stdout = outfile*/

	// убьем ранее запущенный eGaaS
	// kill previously run eGaaS
	if !utils.Mobile() {
		fmt.Println("kill daylight.pid")
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

			sql.DB, err = sql.NewDbConnect()

			err = KillPid(pidMap["pid"])
			if nil != err {
				fmt.Println(err)
				log.Error("KillPid %v", utils.ErrInfo(err))
			}
			if fmt.Sprintf("%s", err) != "null" {
				fmt.Println(fmt.Sprintf("%s", err))
				// даем 15 сек, чтобы завершиться предыдущему процессу
				// give 15 sec to end the previous process
				for i := 0; i < 15; i++ {
					log.Debug("waiting killer %d", i)
					if _, err := os.Stat(*utils.Dir + "/daylight.pid"); err == nil {
						fmt.Println("waiting killer")
						time.Sleep(time.Second)
					} else { // если daylight.pid нет, значит завершился // if there is no daylight.pid, so it is finished
						break
					}
				}
			}
		}
	}

	controllers.SessInit()
	config.MonitorChanges()

	go func() {
		var err error
		sql.DB, err = sql.NewDbConnect()
		log.Debug("%v", sql.DB)
		IosLog("utils.DB:" + fmt.Sprintf("%v", sql.DB))
		if err != nil {
			IosLog("err:" + fmt.Sprintf("%s", utils.ErrInfo(err)))
			log.Error("%v", utils.ErrInfo(err))
			Exit(1)
		}
	}()

	f, err := os.OpenFile(*utils.Dir+"/dclog.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		IosLog("err:" + fmt.Sprintf("%s", utils.ErrInfo(err)))
		log.Error("%v", utils.ErrInfo(err))
		Exit(1)
	}
	defer f.Close()

	if fi, err := os.Stat(*utils.Dir + `/logo.png`); err == nil && fi.Size() > 0 {
		utils.LogoExt = `png`
	}
	IosLog("configIni:" + fmt.Sprintf("%v", config.ConfigIni))
	var backend *logging.LogBackend
	switch config.ConfigIni["log_output"] {
	case "file":
		backend = logging.NewLogBackend(f, "", 0)
	case "console":
		backend = logging.NewLogBackend(os.Stderr, "", 0)
	case "file_console":
	//backend = logging.NewLogBackend(io.MultiWriter(f, os.Stderr), "", 0)
	default:
		backend = logging.NewLogBackend(f, "", 0)
	}
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)

	level := "DEBUG"
	if *utils.LogLevel == "" {
		level = config.ConfigIni["log_level"]
		*utils.LogLevel = level
	} else {
		level = *utils.LogLevel
	}
	logLevel, err := logging.LogLevel(level)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
	}

	log.Error("logLevel: %v", logLevel)
	backendLeveled.SetLevel(logLevel, "")
	logging.SetBackend(backendLeveled)

	rand.Seed(time.Now().UTC().UnixNano())

	// если есть OldFileName, значит работаем под именем dc.tmp и нужно перезапуститься под нормальным именем
	// if there is OldFileName, so act on behalf dc.tmp and we have to restart on behalf the normal name
	log.Debug("OldFileName %v", *utils.OldFileName)
	if *utils.OldFileName != "" || len(configIni) != 0 {

		if *utils.OldFileName != "" { //*utils.Dir+`/dc.tmp`
			err = utils.CopyFileContents(os.Args[0], *utils.OldFileName)
			if err != nil {
				log.Debug("%v", os.Stderr)
				log.Debug("%v", utils.ErrInfo(err))
			}
		}
		// ждем подключения к БД
		// waiting for connection to the database
		for {
			if sql.DB == nil || sql.DB.DB == nil {
				time.Sleep(time.Second)
				continue
			}
			break
		}
		schema.Migration()

		if *utils.OldFileName != "" {
			err = sql.DB.Close()
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			err = os.Remove(filepath.Join(*utils.Dir, "daylight.pid"))
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}

			if thrustWindowLoder != nil {
				thrustWindowLoder.Close()
			}
			system.Finish(0)
			err = exec.Command(*utils.OldFileName, "-dir", *utils.Dir).Start()
			if err != nil {
				log.Debug("%v", os.Stderr)
				log.Debug("%v", utils.ErrInfo(err))
			}
			os.Exit(1)
		}
	}

	// сохраним текущий pid и версию
	// save the current pid and version
	if !utils.Mobile() {
		pid := os.Getpid()
		PidAndVer, err := json.Marshal(map[string]string{"pid": converter.IntToStr(pid), "version": consts.VERSION})
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
		err = ioutil.WriteFile(*utils.Dir+"/daylight.pid", PidAndVer, 0644)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			panic(err)
		}
	}

	// откат БД до указанного блока
	// database rollback to the specified block
	if *utils.RollbackToBlockID > 0 {
		sql.DB, err = sql.NewDbConnect()

		if err := template.LoadContracts(); err != nil {
			log.Error(`Load Contracts`, err)
		}
		parser := new(parser.Parser)
		parser.DCDB = sql.DB
		err = parser.RollbackToBlockID(*utils.RollbackToBlockID)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fmt.Println("complete")
		// получим стату по всем таблам
		// we recieve the statistics of all tables
		allTable, err := sql.DB.GetAllTables()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		startData := map[string]int64{"install": 1, "config": 1, "queue_tx": 99999, "log_transactions": 1, "transactions_status": 99999, "block_chain": 1, "info_block": 1, "dlt_wallets": 1, "confirmations": 9999999, "full_nodes": 1, "system_parameters": 4, "my_node_keys": 99999, "transactions": 999999}
		for _, table := range allTable {
			count, err := sql.DB.Single(`SELECT count(*) FROM ` + converter.EscapeName(table)).Int64()
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			if count > 0 && count > startData[table] {
				fmt.Println(">>ALERT<<", table, count)
			} else {
				fmt.Println(table, "ok")
			}
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

	log.Debug("daemonsStart")
	IosLog("daemonsStart")

	daemons.StartDaemons()

	IosLog("MonitorDaemons")
	// мониторинг демонов
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

	// сигналы демонам для выхода
	// signals for daemons to exit
	IosLog("signals")
	stopdaemons.Signals()

	time.Sleep(time.Second)

	// мониторим сигнал из БД о том, что демонам надо завершаться
	// monitor the signal from the database that the daemons must be completed
	go stopdaemons.WaitStopTime()

	BrowserHTTPHost := "http://localhost:" + *utils.ListenHTTPPort
	HandleHTTPHost := ""
	ListenHTTPHost := *utils.TCPHost + ":" + *utils.ListenHTTPPort
	go func() {
		// уже прошел процесс инсталяции, где юзер указал БД и был перезапуск кошелька
		// The installation process is already finished (where user has specified DB and where wallet has been restarted)
		if len(configIni["db_type"]) > 0 {
			for {
				// ждем, пока произойдет подключение к БД в другой гоурутине
				// wait while connection to a DB in other gourutina takes place
				if sql.DB == nil || sql.DB.DB == nil {
					time.Sleep(time.Second)
					fmt.Println("wait DB")
				} else {
					break
				}
			}
			fmt.Println("GET http host")
			if err := template.LoadContracts(); err != nil {
				log.Error(`Load Contracts`, err)
			}
			BrowserHTTPHost, HandleHTTPHost, ListenHTTPHost = GetHTTPHost()
			// для ноды тоже нужна БД // DB is needed for node as well
			tcpListener()
		}
		IosLog(fmt.Sprintf("BrowserHTTPHost: %v, HandleHTTPHost: %v, ListenHTTPHost: %v", BrowserHTTPHost, HandleHTTPHost, ListenHTTPHost))
		fmt.Printf("BrowserHTTPHost: %v, HandleHTTPHost: %v, ListenHTTPHost: %v\n", BrowserHTTPHost, HandleHTTPHost, ListenHTTPHost)
		go controllers.GetChain()

		route := httprouter.New()
		setRoute(route, `/`, controllers.Index, `GET`)
		setRoute(route, `/content`, controllers.Content, `GET`, `POST`)
		setRoute(route, `/template`, controllers.Template, `GET`, `POST`)
		setRoute(route, `/app`, controllers.App, `GET`, `POST`)
		setRoute(route, `/ajax`, controllers.Ajax, `GET`, `POST`)
		setRoute(route, `/wschain`, controllers.WsBlockchain, `GET`)
		setRoute(route, `/exchangeapi/:name`, exchangeapi.API, `GET`, `POST`)
		api.Route(route)
		route.Handler(`GET`, `/static/*filepath`, http.FileServer(&assetfs.AssetFS{Asset: FileAsset, AssetDir: static.AssetDir, Prefix: ""}))
		route.Handler(`GET`, `/.well-known/*filepath`, http.FileServer(http.Dir(*utils.TLS)))
		if len(*utils.TLS) > 0 {
			go http.ListenAndServeTLS(":443", *utils.TLS+`/fullchain.pem`, *utils.TLS+`/privkey.pem`, route)
		}

		log.Debug("ListenHTTPHost", ListenHTTPHost)

		IosLog(fmt.Sprintf("ListenHTTPHost: %v", ListenHTTPHost))

		fmt.Println("ListenHTTPHost", ListenHTTPHost)

		httpListener(ListenHTTPHost, &BrowserHTTPHost, route)
		// for ipv6 server
		httpListenerV6(route)

		if *utils.Console == 0 && !utils.Mobile() {
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
				openBrowser(BrowserHTTPHost)
			}
		}
	}()

	// ожидает появления свежих записей в чате, затем ждет появления коннектов
	// waits for new records in chat, then waits for connect
	// (заносятся из демеона connections и от тех, кто сам подключился к ноде)
	// (they are entered from the 'connections' daemon and from those who connected to the node by their own)
	// go utils.ChatOutput(utils.ChatNewTx)

	log.Debug("ALL RIGHT")
	IosLog("ALL RIGHT")
	fmt.Println("ALL RIGHT")
	time.Sleep(time.Second * 3600 * 24 * 90)
	log.Debug("EXIT")
}
