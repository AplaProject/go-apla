// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package daylight

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/AplaProject/go-apla/packages/api"
	conf "github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/daemons"
	"github.com/AplaProject/go-apla/packages/daylight/daemonsctl"
	logtools "github.com/AplaProject/go-apla/packages/log"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/modes"
	"github.com/AplaProject/go-apla/packages/network/httpserver"
	"github.com/AplaProject/go-apla/packages/obsmanager"
	"github.com/AplaProject/go-apla/packages/publisher"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/statsd"
	"github.com/AplaProject/go-apla/packages/utils"

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
	switch conf.Config.Log.LogFormat {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.TextFormatter{})
	}
	switch conf.Config.Log.LogTo {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "syslog":
		facility := conf.Config.Log.Syslog.Facility
		tag := conf.Config.Log.Syslog.Tag
		sysLogHook, err := logtools.NewSyslogHook(tag, facility)
		if err != nil {
			log.WithError(err).Error("initializing syslog hook")
		} else {
			log.AddHook(sysLogHook)
		}
	default:
		fileName := filepath.Join(conf.Config.DataDir, conf.Config.Log.LogTo)
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

	switch conf.Config.Log.LogLevel {
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
	log.AddHook(logtools.HexHook{})

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

func initRoutes(listenHost string) {
	handler := modes.RegisterRoutes()
	handler = api.WithCors(handler)
	handler = httpserver.NewMaxBodyReader(handler, conf.Config.HTTPServerMaxBodySize)

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
		go func() {
			err := http.ListenAndServeTLS(listenHost, conf.Config.TLSCert, conf.Config.TLSKey, handler)
			if err != nil {
				log.WithFields(log.Fields{"host": listenHost, "error": err, "type": consts.NetworkError}).Fatal("Listening TLS server")
			}
		}()
		log.WithFields(log.Fields{"host": listenHost}).Info("listening with TLS at")
		return
	} else if len(conf.Config.TLSCert) != 0 || len(conf.Config.TLSKey) != 0 {
		log.Fatal("-tls/TLS must be specified with -tls-cert/TLSCert and -tls-key/TLSKey")
	}

	httpListener(listenHost, handler)
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

	log.WithFields(log.Fields{"mode": conf.Config.OBSMode}).Info("Node running mode")
	if conf.Config.FuncBench {
		log.Warning("Warning! Access checking is disabled in some built-in functions")
	}

	f := utils.LockOrDie(conf.Config.LockFilePath)
	defer f.Unlock()

	if err := utils.MakeDirectory(conf.Config.TempDir); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.IOError, "dir": conf.Config.TempDir}).Error("can't create temporary directory")
		Exit(1)
	}

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

	smart.InitVM()
	if err := syspar.ReadNodeKeys(); err != nil {
		log.Errorf("can't read node keys: %s", err)
		Exit(1)
	}
	if model.DBConn != nil {
		if err := model.UpdateSchema(); err != nil {
			log.WithFields(log.Fields{"error": err}).Error("on running update migrations")
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		utils.CancelFunc = cancel
		utils.ReturnCh = make(chan string)

		// The installation process is already finished (where user has specified DB and where wallet has been restarted)
		err := daemonsctl.RunAllDaemons(ctx)
		log.Info("Daemons started")
		if err != nil {
			os.Exit(1)
		}

		obsmanager.InitOBSManager()
	}
	daemons.WaitForSignals()

	initRoutes(conf.Config.HTTP.Str())

	select {}
}
