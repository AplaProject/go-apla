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
	_ "image/png"
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

	// читаем config.ini
	configIni := make(map[string]string)
	configIni_, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
	if err != nil {
		IosLog("err:" + fmt.Sprintf("%s", utils.ErrInfo(err)))
		log.Error("%v", utils.ErrInfo(err))
	} else {
		configIni, err = configIni_.GetSection("default")
	}

	if *utils.TcpHost == "" {
		*utils.TcpHost = configIni["tcp_host"]
	}
	if *utils.FirstBlockDir == "" {
		*utils.FirstBlockDir = configIni["first_block_dir"]
	}
	if *utils.ListenHttpPort == "" {
		*utils.ListenHttpPort = configIni["http_port"]
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
				for i := 0; i < 15; i++ {
					log.Debug("waiting killer %d", i)
					if _, err := os.Stat(*utils.Dir + "/daylight.pid"); err == nil {
						fmt.Println("waiting killer")
						utils.Sleep(1)
					} else { // если daylight.pid нет, значит завершился
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

	logLevel_ := "DEBUG"
	if *utils.LogLevel == "" {
		logLevel_ = configIni["log_level"]
		*utils.LogLevel = logLevel_
	} else {
		logLevel_ = *utils.LogLevel
	}

	logLevel, err := logging.LogLevel(logLevel_)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
	}

	log.Error("logLevel: %v", logLevel)
	backendLeveled.SetLevel(logLevel, "")
	logging.SetBackend(backendLeveled)

	rand.Seed(time.Now().UTC().UnixNano())

	// если есть OldFileName, значит работаем под именем dc.tmp и нужно перезапуститься под нормальным именем
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
	if *utils.RollbackToBlockId > 0 {
		utils.DB, err = utils.NewDbConnect(configIni)

		if err := utils.LoadContracts(); err != nil {
			log.Error(`Load Contracts`, err)
		}
		parser := new(parser.Parser)
		parser.DCDB = utils.DB
		err = parser.RollbackToBlockId(*utils.RollbackToBlockId)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fmt.Println("complete")
		// получим стату по всем таблам
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
	go stopdaemons.WaitStopTime()

	BrowserHttpHost := "http://localhost:" + *utils.ListenHttpPort
	HandleHttpHost := ""
	ListenHttpHost := *utils.TcpHost + ":" + *utils.ListenHttpPort
	go func() {
		// уже прошли процесс инсталяции, где юзер указал БД и был перезапуск кошелька
		if len(configIni["db_type"]) > 0 {
			for {
				// ждем, пока произойдет подключение к БД в другой гоурутине
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
			BrowserHttpHost, HandleHttpHost, ListenHttpHost = GetHttpHost()
			// для ноды тоже нужна БД
			tcpListener()
		}
		IosLog(fmt.Sprintf("BrowserHttpHost: %v, HandleHttpHost: %v, ListenHttpHost: %v", BrowserHttpHost, HandleHttpHost, ListenHttpHost))
		fmt.Printf("BrowserHttpHost: %v, HandleHttpHost: %v, ListenHttpHost: %v\n", BrowserHttpHost, HandleHttpHost, ListenHttpHost)
		go controllers.GetChain()
		// включаем листинг веб-сервером для клиентской части
		http.HandleFunc(HandleHttpHost+"/", controllers.Index)
		http.HandleFunc(HandleHttpHost+"/content", controllers.Content)
		http.HandleFunc(HandleHttpHost+"/template", controllers.Template)
		http.HandleFunc(HandleHttpHost+"/app", controllers.App)
		http.HandleFunc(HandleHttpHost+"/ajax", controllers.Ajax)
		http.HandleFunc(HandleHttpHost+"/wschain", controllers.WsBlockchain)
		http.HandleFunc(HandleHttpHost+"/exchangeapi/newkey", exchangeapi.Api)
		http.HandleFunc(HandleHttpHost+"/exchangeapi/send", exchangeapi.Api)
		http.HandleFunc(HandleHttpHost+"/exchangeapi/balance", exchangeapi.Api)
		http.HandleFunc(HandleHttpHost+"/exchangeapi/history", exchangeapi.Api)
		//http.HandleFunc(HandleHttpHost+"/ajaxjson", controllers.AjaxJson)
		//http.HandleFunc(HandleHttpHost+"/tools", controllers.Tools)
		//http.Handle(HandleHttpHost+"/public/", noDirListing(http.FileServer(http.Dir(*utils.Dir))))
		http.Handle(HandleHttpHost+"/static/", http.FileServer(&assetfs.AssetFS{Asset: FileAsset, AssetDir: static.AssetDir, Prefix: ""}))
		if len(*utils.Tls) > 0 {
			http.Handle(HandleHttpHost+"/.well-known/", http.FileServer(http.Dir(*utils.Tls)))
			httpsMux := http.NewServeMux()
			httpsMux.HandleFunc(HandleHttpHost+"/", controllers.Index)
			httpsMux.HandleFunc(HandleHttpHost+"/content", controllers.Content)
			httpsMux.HandleFunc(HandleHttpHost+"/ajax", controllers.Ajax)
			httpsMux.HandleFunc(HandleHttpHost+"/wschain", controllers.WsBlockchain)
			httpsMux.HandleFunc(HandleHttpHost+"/exchangeapi/newkey", exchangeapi.Api)
			httpsMux.HandleFunc(HandleHttpHost+"/exchangeapi/send", exchangeapi.Api)
			httpsMux.HandleFunc(HandleHttpHost+"/exchangeapi/balance", exchangeapi.Api)
			httpsMux.HandleFunc(HandleHttpHost+"/exchangeapi/history", exchangeapi.Api)
			httpsMux.Handle(HandleHttpHost+"/static/", http.FileServer(&assetfs.AssetFS{Asset: FileAsset, AssetDir: static.AssetDir, Prefix: ""}))
			go http.ListenAndServeTLS(":443", *utils.Tls+`/fullchain.pem`, *utils.Tls+`/privkey.pem`, httpsMux)
		}

		log.Debug("ListenHttpHost", ListenHttpHost)

		IosLog(fmt.Sprintf("ListenHttpHost: %v", ListenHttpHost))

		fmt.Println("ListenHttpHost", ListenHttpHost)

		httpListener(ListenHttpHost, &BrowserHttpHost)
		// for ipv6 server
		httpListenerV6()

		if *utils.Console == 0 && !utils.Mobile() {
			utils.Sleep(1)
			if thrustWindowLoder != nil {
				thrustWindowLoder.Close()
				thrustWindow := thrust.NewWindow(thrust.WindowOptions{
					RootUrl: BrowserHttpHost,
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
				openBrowser(BrowserHttpHost)
			}
		}
	}()

	// ожидает появления свежих записей в чате, затем ждет появления коннектов
	// (заносятся из демеона connections и от тех, кто сам подключился к ноде)
	//go utils.ChatOutput(utils.ChatNewTx)

	log.Debug("ALL RIGHT")
	IosLog("ALL RIGHT")
	fmt.Println("ALL RIGHT")
	utils.Sleep(3600 * 24 * 90)
	log.Debug("EXIT")
}
