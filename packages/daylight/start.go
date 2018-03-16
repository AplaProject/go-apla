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
	"path/filepath"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/api"
	conf "github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/daemons"
	"github.com/GenesisKernel/go-genesis/packages/daylight/daemonsctl"
	logtools "github.com/GenesisKernel/go-genesis/packages/log"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/publisher"
	"github.com/GenesisKernel/go-genesis/packages/statsd"
	"github.com/GenesisKernel/go-genesis/packages/utils"

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
	pidPath := conf.Config.GetPidPath()
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
		log.WithFields(log.Fields{"path": conf.Config.DataDir + pidMap["pid"]}).Debug("old pid path")

		KillPid(pidMap["pid"])
		if fmt.Sprintf("%s", err) != "null" {
			// give 15 sec to end the previous process
			for i := 0; i < 15; i++ {
				if _, err := os.Stat(conf.Config.GetPidPath()); err == nil {
					time.Sleep(time.Second)
				} else {
					break
				}
			}
		}
	}
}

func initLogs() error {
	if len(conf.Config.LogConfig.LogTo) == 0 {
		log.SetOutput(os.Stdout)
	} else {
		fileName := filepath.Join(conf.Config.DataDir, conf.Config.LogConfig.LogTo)
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

	switch conf.Config.LogConfig.LogLevel {
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

	return ioutil.WriteFile(conf.Config.GetPidPath(), PidAndVer, 0644)
}

func delPidFile() {
	os.Remove(conf.Config.GetPidPath())
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
	if conf.Config.TLS {
		if len(conf.Config.TLSCert) == 0 || len(conf.Config.TLSKey) == 0 {
			log.Fatal("-tls-cert/TLSCert and -tls-key/TLSKey must be specified with -tls/TLS")
		}
		if _, err := os.Stat(conf.Config.TLSCert); os.IsNotExist(err) {
			log.WithError(err).Fatalf(`Filepath -tls-cert/TLSCert = %s is invalid`, conf.Config.TLSCert)
		}
		if _, err := os.Stat(conf.Config.TLSKey); os.IsNotExist(err) {
			log.WithError(err).Fatalf(`Filepath -tls-key/TLSKey = %s is invalid`, conf.Config.TLSKey)
		}
		go http.ListenAndServeTLS(":443", conf.Config.TLSCert, conf.Config.TLSKey, route)
	} else if len(conf.Config.TLSCert) != 0 || len(conf.Config.TLSKey) != 0 {
		log.Fatal("-tls/TLS must be specified with -tls-cert/TLSCert and -tls-key/TLSKey")
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

	f := utils.LockOrDie(conf.Config.LockFilePath)
	defer f.Unlock()

	conf.Config.Installed = true

	initGorm(conf.Config.DB)
	log.WithFields(log.Fields{"work_dir": conf.Config.DataDir, "version": consts.VERSION}).Info("started with")

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
