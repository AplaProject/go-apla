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
	"log/syslog"
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

func syslogFacility(facility string) syslog.Priority {
	switch facility {
	case "LOG_KERN":
		return syslog.LOG_KERN
	case "LOG_USER":
		return syslog.LOG_USER
	case "LOG_MAIL":
		return syslog.LOG_MAIL
	case "LOG_DAEMON":
		return syslog.LOG_DAEMON
	case "LOG_AUTH":
		return syslog.LOG_AUTH
	case "LOG_SYSLOG":
		return syslog.LOG_SYSLOG
	case "LOG_LPR":
		return syslog.LOG_LPR
	case "LOG_NEWS":
		return syslog.LOG_NEWS
	case "LOG_UUCP":
		return syslog.LOG_UUCP
	case "LOG_CRON":
		return syslog.LOG_CRON
	case "LOG_AUTHPRIV":
		return syslog.LOG_AUTHPRIV
	case "LOG_FTP":
		return syslog.LOG_FTP
	case "LOG_LOCAL0":
		return syslog.LOG_LOCAL0
	case "LOG_LOCAL1":
		return syslog.LOG_LOCAL1
	case "LOG_LOCAL2":
		return syslog.LOG_LOCAL2
	case "LOG_LOCAL3":
		return syslog.LOG_LOCAL3
	case "LOG_LOCAL4":
		return syslog.LOG_LOCAL4
	case "LOG_LOCAL5":
		return syslog.LOG_LOCAL5
	case "LOG_LOCAL6":
		return syslog.LOG_LOCAL6
	case "LOG_LOCAL7":
		return syslog.LOG_LOCAL7
	}
	return 0
}

func syslogSeverity(severity string) syslog.Priority {
	switch severity {
	case "LOG_EMERG":
		return syslog.LOG_EMERG
	case "LOG_ALERT":
		return syslog.LOG_ALERT
	case "LOG_CRIT":
		return syslog.LOG_CRIT
	case "LOG_ERR":
		return syslog.LOG_ERR
	case "LOG_WARNING":
		return syslog.LOG_WARNING
	case "LOG_NOTICE":
		return syslog.LOG_NOTICE
	case "LOG_INFO":
		return syslog.LOG_INFO
	case "LOG_DEBUG":
		return syslog.LOG_DEBUG
	}
	return 0
}

func initLogs() error {
	switch conf.Config.LogConfig.LogFormat {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.TextFormatter{})
	}
	switch conf.Config.LogConfig.LogTo {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "syslog":
		severity := syslogSeverity(conf.Config.LogConfig.Syslog.Severity)
		facility := syslogFacility(conf.Config.LogConfig.Syslog.Facility)
		tag := conf.Config.LogConfig.Syslog.Tag
		sysLogHook, err := logtools.NewSyslogHook(tag, severity|facility)
		if err != nil {
			log.WithError(err).Error("initializing syslog hook")
		} else {
			log.AddHook(sysLogHook)
		}
	default:
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

func CreateLockFile() error {
	return ioutil.WriteFile(conf.Config.LockFilePath, []byte{}, 0644)
}

func delPidFile() {
	os.Remove(conf.Config.GetPidPath())
}

func DelLockFile() error {
	return os.Remove(conf.Config.LockFilePath)
}

func IsLockFileExists() bool {
	if _, err := os.Stat(conf.Config.LockFilePath); os.IsNotExist(err) {
		return false
	}

	return true
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
	route.Handler(`GET`, consts.WellKnownRoute, http.FileServer(http.Dir(conf.Config.TLS)))
	if len(conf.Config.TLS) > 0 {
		go http.ListenAndServeTLS(":443", conf.Config.TLS+consts.TLSFullchainPem, conf.Config.TLS+consts.TLSPrivkeyPem, route)
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

	if IsLockFileExists() {
		log.Fatal("Lock file is found")
	}

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

	// create lock file
	if err := CreateLockFile(); err != nil {
		log.Errorf("can't create lock: %s", err)
		Exit(1)
	}
	defer DelLockFile()

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
