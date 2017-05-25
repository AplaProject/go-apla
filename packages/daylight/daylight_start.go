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

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/controllers"
	"github.com/EGaaS/go-egaas-mvp/packages/daemons"
	"github.com/EGaaS/go-egaas-mvp/packages/exchangeapi"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/schema"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/stopdaemons"
	"github.com/EGaaS/go-egaas-mvp/packages/system"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/astaxie/beego/config"
	"github.com/go-bindata-assetfs"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/thrust"
	"github.com/op/go-logging"
)

func FileAsset(name string) ([]byte, error) {

	if name := strings.Replace(name, "\\", "/", -1); name == `static/img/logo.`+utils.LogoExt {
		logofile := *utils.Dir + `/logo.` + utils.LogoExt
		if fi, err := os.Stat(logofile); err == nil && fi.Size() > 0 {
			return ioutil.ReadFile(logofile)
		}
	}
	return static.Asset(name)
}

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
	configIni := make(map[string]string)
	fullConfigIni, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
	if err != nil {
		IosLog("err:" + fmt.Sprintf("%s", utils.ErrInfo(err)))
		log.Error("%v", utils.ErrInfo(err))
	} else {
		configIni, err = fullConfigIni.GetSection("default")
	}

	if *utils.TCPHost == "" {
		*utils.TCPHost = configIni["tcp_host"]
	}
	if *utils.FirstBlockDir == "" {
		*utils.FirstBlockDir = configIni["first_block_dir"]
	}
	if *utils.ListenHTTPPort == "" {
		*utils.ListenHTTPPort = configIni["http_port"]
	}
	if *utils.Dir == "" {
		*utils.Dir = configIni["dir"]
	}
	utils.OneCountry = utils.StrToInt64(configIni["one_country"])
	utils.PrivCountry = configIni["priv_country"] == `1` || configIni["priv_country"] == `true`
	if len(configIni["lang"]) > 0 {
		utils.LangList = strings.Split(configIni["lang"], `,`)
	}
	/*	outfile, err := os.Create("./out.txt")
	    if err != nil {
	        panic(err)
	    }
	    defer outfile.Close()
		os.Stdout = outfile*/

	// убьем ранее запущенный daylight
	// kill previously run daylight
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

			utils.DB, err = utils.NewDbConnect(configIni)

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
						utils.Sleep(1)
					} else { // если daylight.pid нет, значит завершился // if there is no daylight.pid, so it is finished
						break
					}
				}
			}
		}
	}

	controllers.SessInit()
	controllers.ConfigInit()
	daemons.ConfigInit()

	go func() {
		var err error
		utils.DB, err = utils.NewDbConnect(configIni)
		log.Debug("%v", utils.DB)
		IosLog("utils.DB:" + fmt.Sprintf("%v", utils.DB))
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
	IosLog("configIni:" + fmt.Sprintf("%v", configIni))
	var backend *logging.LogBackend
	switch configIni["log_output"] {
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
		level = configIni["log_level"]
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
			if utils.DB == nil || utils.DB.DB == nil {
				utils.Sleep(1)
				continue
			}
			break
		}
		schema.Migration()

		if *utils.OldFileName != "" {
			err = utils.DB.Close()
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
		PidAndVer, err := json.Marshal(map[string]string{"pid": utils.IntToStr(pid), "version": consts.VERSION})
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
		utils.DB, err = utils.NewDbConnect(configIni)

		if err := utils.LoadContracts(); err != nil {
			log.Error(`Load Contracts`, err)
		}
		parser := new(parser.Parser)
		parser.DCDB = utils.DB
		err = parser.RollbackToBlockID(*utils.RollbackToBlockID)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fmt.Println("complete")
		// получим стату по всем таблам
		// we recieve the statistics of all tables
		allTable, err := utils.DB.GetAllTables()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		startData := map[string]int64{"install": 1, "config": 1, "queue_tx": 99999, "log_transactions": 1, "transactions_status": 99999, "block_chain": 1, "info_block": 1, "dlt_wallets": 1, "confirmations": 9999999, "full_nodes": 1, "system_parameters": 4, "my_node_keys": 99999, "transactions": 999999}
		for _, table := range allTable {
			count, err := utils.DB.Single(`SELECT count(*) FROM ` + lib.EscapeName(table)).Int64()
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
			if utils.Time()%10 == 0 {
				log.Debug("daemonsTable: %v\n", daemonsTable)
			}
		}
	}()

	// сигналы демонам для выхода
	IosLog("signals")
	stopdaemons.Signals()

	utils.Sleep(1)

	// мониторим сигнал из БД о том, что демонам надо завершаться
	// monitor the signal from the database that the daemons must be completed
	go stopdaemons.WaitStopTime()

	BrowserHTTPHost := "http://localhost:" + *utils.ListenHTTPPort
	HandleHTTPHost := ""
	ListenHTTPHost := *utils.TCPHost + ":" + *utils.ListenHTTPPort
	go func() {
		// уже прошли процесс инсталяции, где юзер указал БД и был перезапуск кошелька
		// The installation process is already finished (where user has specified DB and where wallet has been restarted)
		if len(configIni["db_type"]) > 0 {
			for {
				// ждем, пока произойдет подключение к БД в другой гоурутине
				// wait while connection to a DB in other gourutina takes place
				if utils.DB == nil || utils.DB.DB == nil {
					utils.Sleep(1)
					fmt.Println("wait DB")
				} else {
					break
				}
			}
			fmt.Println("GET http host")
			if err := utils.LoadContracts(); err != nil {
				log.Error(`Load Contracts`, err)
			}
			BrowserHTTPHost, HandleHTTPHost, ListenHTTPHost = GetHTTPHost()
			// для ноды тоже нужна БД // DB is needed for node as well
			tcpListener()
		}
		IosLog(fmt.Sprintf("BrowserHTTPHost: %v, HandleHTTPHost: %v, ListenHTTPHost: %v", BrowserHTTPHost, HandleHTTPHost, ListenHTTPHost))
		fmt.Printf("BrowserHTTPHost: %v, HandleHTTPHost: %v, ListenHTTPHost: %v\n", BrowserHTTPHost, HandleHTTPHost, ListenHTTPHost)
		go controllers.GetChain()
		// включаем листинг веб-сервером для клиентской части
		// switch on the listing by web-server for client part
		http.HandleFunc(HandleHTTPHost+"/", controllers.Index)
		http.HandleFunc(HandleHTTPHost+"/content", controllers.Content)
		http.HandleFunc(HandleHTTPHost+"/template", controllers.Template)
		http.HandleFunc(HandleHTTPHost+"/app", controllers.App)
		http.HandleFunc(HandleHTTPHost+"/ajax", controllers.Ajax)
		http.HandleFunc(HandleHTTPHost+"/wschain", controllers.WsBlockchain)
		http.HandleFunc(HandleHTTPHost+"/exchangeapi/newkey", exchangeapi.API)
		http.HandleFunc(HandleHTTPHost+"/exchangeapi/send", exchangeapi.API)
		http.HandleFunc(HandleHTTPHost+"/exchangeapi/balance", exchangeapi.API)
		http.HandleFunc(HandleHTTPHost+"/exchangeapi/history", exchangeapi.API)
		//http.HandleFunc(HandleHTTPHost+"/ajaxjson", controllers.AjaxJson)
		//http.HandleFunc(HandleHTTPHost+"/tools", controllers.Tools)
		//http.Handle(HandleHTTPHost+"/public/", noDirListing(http.FileServer(http.Dir(*utils.Dir))))
		http.Handle(HandleHTTPHost+"/static/", http.FileServer(&assetfs.AssetFS{Asset: FileAsset, AssetDir: static.AssetDir, Prefix: ""}))
		if len(*utils.TLS) > 0 {
			http.Handle(HandleHTTPHost+"/.well-known/", http.FileServer(http.Dir(*utils.TLS)))
			httpsMux := http.NewServeMux()
			httpsMux.HandleFunc(HandleHTTPHost+"/", controllers.Index)
			httpsMux.HandleFunc(HandleHTTPHost+"/content", controllers.Content)
			httpsMux.HandleFunc(HandleHTTPHost+"/ajax", controllers.Ajax)
			httpsMux.HandleFunc(HandleHTTPHost+"/wschain", controllers.WsBlockchain)
			httpsMux.HandleFunc(HandleHTTPHost+"/exchangeapi/newkey", exchangeapi.API)
			httpsMux.HandleFunc(HandleHTTPHost+"/exchangeapi/send", exchangeapi.API)
			httpsMux.HandleFunc(HandleHTTPHost+"/exchangeapi/balance", exchangeapi.API)
			httpsMux.HandleFunc(HandleHTTPHost+"/exchangeapi/history", exchangeapi.API)
			httpsMux.Handle(HandleHTTPHost+"/static/", http.FileServer(&assetfs.AssetFS{Asset: FileAsset, AssetDir: static.AssetDir, Prefix: ""}))
			go http.ListenAndServeTLS(":443", *utils.TLS+`/fullchain.pem`, *utils.TLS+`/privkey.pem`, httpsMux)
		}

		log.Debug("ListenHTTPHost", ListenHTTPHost)

		IosLog(fmt.Sprintf("ListenHTTPHost: %v", ListenHTTPHost))

		fmt.Println("ListenHTTPHost", ListenHTTPHost)

		httpListener(ListenHTTPHost, &BrowserHTTPHost)
		// for ipv6 server
		httpListenerV6()

		if *utils.Console == 0 && !utils.Mobile() {
			utils.Sleep(1)
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
	//go utils.ChatOutput(utils.ChatNewTx)

	log.Debug("ALL RIGHT")
	IosLog("ALL RIGHT")
	fmt.Println("ALL RIGHT")
	utils.Sleep(3600 * 24 * 90)
	log.Debug("EXIT")
}
