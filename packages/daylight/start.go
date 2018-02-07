// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package daylight

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/GenesisCommunity/go-genesis/packages/api"
	"github.com/GenesisCommunity/go-genesis/packages/autoupdate"
	conf "github.com/GenesisCommunity/go-genesis/packages/conf"
	"github.com/GenesisCommunity/go-genesis/packages/config/syspar"
	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/daemons"
	"github.com/GenesisCommunity/go-genesis/packages/daylight/daemonsctl"
	"github.com/GenesisCommunity/go-genesis/packages/install"
	logtools "github.com/GenesisCommunity/go-genesis/packages/log"
	"github.com/GenesisCommunity/go-genesis/packages/model"
	"github.com/GenesisCommunity/go-genesis/packages/parser"
	"github.com/GenesisCommunity/go-genesis/packages/publisher"
	"github.com/GenesisCommunity/go-genesis/packages/smart"
	"github.com/GenesisCommunity/go-genesis/packages/statsd"
	"github.com/GenesisCommunity/go-genesis/packages/utils"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

func initStatsd() {
	cfg := conf.Config.StatsD
	if err := statsd.Init(cfg.Host, cfg.Port, cfg.Name); err != nil {
		log.WithFields(log.Fields{"type": consts.StatsdError, "error": err}).Fatal("cannot initialize statsd")
	}
}

func killOld() {
	pidPath := conf.GetPidFile()
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
		log.WithFields(log.Fields{"path": conf.Config.WorkDir + pidMap["pid"]}).Debug("old pid path")

		KillPid(pidMap["pid"])
		if fmt.Sprintf("%s", err) != "null" {
			// give 15 sec to end the previous process
			for i := 0; i < 15; i++ {
				if _, err := os.Stat(conf.GetPidFile()); err == nil {
					time.Sleep(time.Second)
				} else {
					break
				}
			}
		}
	}
}

func initLogs() error {

	if len(conf.Config.LogFileName) == 0 {
		log.SetOutput(os.Stdout)
	} else {
		fileName := filepath.Join(conf.Config.WorkDir, conf.Config.LogFileName)
		openMode := os.O_APPEND
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			openMode = os.O_CREATE
		}

		f, err := os.OpenFile(fileName, os.O_WRONLY|openMode, 0755)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Can't open log file: ", fileName)
			return err
		}
		log.SetOutput(f)
	}

	switch conf.Config.LogLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	log.AddHook(logtools.ContextHook{})

	return nil
}

func savePid() error {
	pid := os.Getpid()
	PidAndVer, err := json.Marshal(map[string]string{"pid": converter.IntToStr(pid), "version": consts.VERSION})
	if err != nil {
		log.WithFields(log.Fields{"pid": pid, "error": err, "type": consts.JSONMarshallError}).Error("marshalling pid to json")
		return err
	}
	return ioutil.WriteFile(conf.GetPidFile(), PidAndVer, 0644)
}

func delPidFile() {
	os.Remove(conf.GetPidFile())
}

func rollbackToBlock(blockID int64) error {
	if err := smart.LoadContracts(nil); err != nil {
		return err
	}
	parser := new(parser.Parser)
	err := parser.RollbackToBlockID(*conf.RollbackToBlockID)
	if err != nil {
		return err
	}

	// block id = 1, is a special case for full rollback
	if blockID != 1 {
		return nil
	}

	// check blocks related tables
	startData := map[string]int64{"1_menu": 1, "1_pages": 1, "1_contracts": 26, "1_parameters": 11, "1_keys": 1, "1_tables": 8, "stop_daemons": 1, "queue_blocks": 9999999, "system_tables": 1, "system_parameters": 27, "system_states": 1, "install": 1, "queue_tx": 9999999, "log_transactions": 1, "transactions_status": 9999999, "block_chain": 1, "info_block": 1, "confirmations": 9999999, "transactions": 9999999}
	warn := 0
	for table := range startData {
		count, err := model.GetRecordsCountTx(nil, table)
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
		rbFile := filepath.Join(conf.Config.WorkDir, consts.RollbackResultFilename)
		ioutil.WriteFile(rbFile, []byte("1"), 0644)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.WritingFile, "path": rbFile}).Error("rollback result flag")
			return err
		}
	}
	return nil
}

func setRoute(route *httprouter.Router, path string, handle func(http.ResponseWriter, *http.Request), methods ...string) {
	for _, method := range methods {
		route.HandlerFunc(method, path, handle)
	}
}

func initRoutes(listenHost string) {
	route := httprouter.New()
	setRoute(route, `/monitoring`, daemons.Monitoring, `GET`)
	api.Route(route)
	route.Handler(`GET`, consts.WellKnownRoute, http.FileServer(http.Dir(*conf.TLS)))
	if len(*conf.TLS) > 0 {
		go http.ListenAndServeTLS(":443", *conf.TLS+consts.TLSFullchainPem, *conf.TLS+consts.TLSPrivkeyPem, route)
	}

	httpListener(listenHost, route)
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
		delPidFile()
		model.GormClose()
		statsd.Close()
		os.Exit(code)
	}

	initGorm := func(dbCfg conf.DBConfig) {
		err = model.GormInit(dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.Name)
		if err != nil {
			log.WithFields(log.Fields{
				"db_user": dbCfg.User, "db_password": dbCfg.Password, "db_name": dbCfg.Name, "type": consts.DBError,
			}).Error("can't init gorm")
			Exit(1)
		}
	}

	conf.InitConfigFlags()
	if conf.NoConfig() {
		conf.Installed = false
		log.Info("Config file missing.")
	} else {
		if !*conf.InitConfig {
			if err := conf.LoadConfig(); err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("LoadConfig")
				return
			}
			conf.Installed = true
		}
	}
	conf.SetConfigParams()

	autoupdate.InitUpdater(conf.Config.Autoupdate.ServerAddress, conf.Config.Autoupdate.PublicKeyPath)

	// process directives
	if *conf.GenerateFirstBlock {
		if err := install.GenerateFirstBlock(); err != nil {
			log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("GenerateFirstBlock")
			Exit(1)
		}
	}

	if *conf.InitDatabase {
		if err := model.InitDB(conf.Config.DB); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("InitDB")
			Exit(1)
		}
	}

	if *conf.InitConfig {
		if err := conf.SaveConfig(); err != nil {
			log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("Error writing config file")
			Exit(1)
		}
		log.Info("Config file created.")
		conf.Installed = true
	}

	if conf.Installed {
		if conf.Config.KeyID == 0 {
			key, err := parser.GetKeyIDFromPrivateKey()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("Unable to get KeyID")
				Exit(1)
			}
			conf.Config.KeyID = key
			if err := conf.SaveConfig(); err != nil {
				log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("Error writing config file")
				Exit(1)
			}
		}
		initGorm(conf.Config.DB)

		err = autoupdate.Run()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.AutoupdateError, "error": err}).Error("run autoupdate")
		}
	}

	log.WithFields(log.Fields{"work_dir": conf.Config.WorkDir, "version": consts.VERSION}).Info("started with")

	killOld()

	publisher.InitCentrifugo(conf.Config.Centrifugo)

	initStatsd()

	err = initLogs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "logs init failed: %v\n", utils.ErrInfo(err))
		Exit(1)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	// save the current pid and version
	if err := savePid(); err != nil {
		log.Errorf("can't create pid: %s", err)
		Exit(1)
	}
	defer delPidFile()

	// database rollback to the specified block
	if *conf.RollbackToBlockID > 0 {
		err = syspar.SysUpdate(nil)
		if err != nil {
			log.WithError(err).Error("can't read system parameters")
		}
		log.WithFields(log.Fields{"block_id": *conf.RollbackToBlockID}).Info("Rollbacking to block ID")
		err := rollbackToBlock(*conf.RollbackToBlockID)
		log.WithFields(log.Fields{"block_id": *conf.RollbackToBlockID}).Info("Rollback is ok")
		if err != nil {
			log.WithError(err).Error("Rollback error")
		} else {
			log.Info("Rollback is OK")
		}
		Exit(0)
	}

	if *conf.NoStart {
		Exit(0)
	}

	if model.DBConn != nil {
		// The installation process is already finished (where user has specified DB and where wallet has been restarted)
		err := daemonsctl.RunAllDaemons()
		log.Info("Daemons started")
		if err != nil {
			os.Exit(1)
		}
	}

	daemons.WaitForSignals()

	initRoutes(conf.Config.HTTP.Str())

	select {}
}
