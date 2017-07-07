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

package daemons

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	"github.com/astaxie/beego/config"
	"github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("daemons")
	/*DaemonCh        chan bool     = make(chan bool, 100)
	AnswerDaemonCh  chan string   = make(chan string, 100)*/

	// MonitorDaemonCh is a channel for daemons
	MonitorDaemonCh = make(chan []string, 100)
	configIni       map[string]string
)

type daemon struct {
	*sql.DCDB
	goRoutineName string
	/*DaemonCh       chan bool
	AnswerDaemonCh chan string*/
	sleepTime time.Duration
}

// ConfigInit regularly reads config.ini file
func ConfigInit() {
	// мониторим config.ini на наличие изменений
	// monitor config.ini for changes
	go func() {
		for {
			logger.Debug("ConfigInit monitor")
			if _, err := os.Stat(*utils.Dir + "/config.ini"); os.IsNotExist(err) {
				time.Sleep(time.Second)
				continue
			}
			confIni, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
			if err != nil {
				logger.Error("%v", utils.ErrInfo(err))
			}
			configIni, err = confIni.GetSection("default")
			if err != nil {
				logger.Error("%v", utils.ErrInfo(err))
			}
			if len(configIni["db_type"]) > 0 {
				break
			}
			time.Sleep(time.Second * 3)
		}
	}()
}

func init() {
	flag.Parse()

}

var newDaemonsList = map[string]func(*daemon, context.Context) error{
	"CreatingBlockchain": CreatingBlockchain,
	"Disseminator":       Disseminator,
	"BlockGenerator":     BlockGenerator,
	"QueueParserTx":      QueueParserTx,
	"QueueParserBlocks":  QueueParserBlocks,
	"Confirmations":      Confirmations,
	"BlocksCollection":   BlocksCollection,
	"UpdFullNodes":       UpdFullNodes,
}

var serverList = []string{
	"CreatingBlockchain",
	"BlockGenerator",
	"QueueParserTx",
	"QueueParserBlocks",
	"Disseminator",
	"Confirmations",
	"BlocksCollection",
	"UpdFullNodes",
}

var mobileList = []string{
	"QueueParserTx",
	"Disseminator",
	"Confirmations",
	"BlocksCollection",
}

var rollbackList = []string{
	"BlocksCollection",
	"Confirmations",
}

func daemonLoop(ctx context.Context, goRoutineName string, handler func(*daemon, context.Context) error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	db, err := sql.WaitDB(ctx)
	if err != nil {
		return
	}

	d := &daemon{
		DCDB:          db,
		goRoutineName: goRoutineName,
		sleepTime:     1,
	}

	timer := time.NewTimer(time.Duration(d.sleepTime) * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-timer.C:
			logger.Info(d.goRoutineName)
			MonitorDaemonCh <- []string{d.goRoutineName, converter.Int64ToStr(time.Now().Unix())}

			err = handler(d, ctx)
			if err != nil {
				logger.Errorf("confirmation error %s", err)
			}
			timer.Reset(time.Duration(d.sleepTime) * time.Second)
		}
	}
}

// StartDaemons starts daemons
func StartDaemons() {
	utils.DaemonsChans = nil

	if configIni["daemons"] == "null" {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	utils.CancelFunc = cancel

	daemonsToStart := serverList
	if utils.Mobile() {
		daemonsToStart = mobileList
	}
	if *utils.TestRollBack == 1 {
		daemonsToStart = rollbackList
	}

	if len(configIni["daemons"]) > 0 {
		daemonsToStart = strings.Split(configIni["daemons"], ",")
	}

	for _, name := range daemonsToStart {
		handler, ok := newDaemonsList[name]
		if ok {
			go daemonLoop(ctx, name, handler)
			continue
		}

		logger.Errorf("unknown daemon name: %s", name)

	}
}
