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
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("daemons")
	/*DaemonCh        chan bool     = make(chan bool, 100)
	AnswerDaemonCh  chan string   = make(chan string, 100)*/

	// MonitorDaemonCh is a channel for daemons
	MonitorDaemonCh = make(chan []string, 100)
)

type daemon struct {
	goRoutineName string
	/*DaemonCh       chan bool
	AnswerDaemonCh chan string*/
	sleepTime time.Duration
}

func init() {
	flag.Parse()
}

var daemonsList = map[string]func(*daemon, context.Context) error{
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

func daemonLoop(ctx context.Context, goRoutineName string, handler func(*daemon, context.Context) error, retCh chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	err := WaitDB(ctx)
	if err != nil {
		return
	}

	d := &daemon{
		goRoutineName: goRoutineName,
		sleepTime:     1 * time.Second,
	}

	err = handler(d, ctx)
	if err != nil {
		logger.Errorf("daemon %s error: %s (%v)", goRoutineName, err, utils.Caller(1))
	}

	for {
		select {
		case <-ctx.Done():
			retCh <- goRoutineName
			return

		case <-time.After(d.sleepTime):
			logger.Info(d.goRoutineName)
			MonitorDaemonCh <- []string{d.goRoutineName, converter.Int64ToStr(time.Now().Unix())}

			err = handler(d, ctx)
			if err != nil {
				logger.Errorf("daemon %s error: %s (%v)", goRoutineName, err, utils.Caller(2))
			}

		}
	}
}

// StartDaemons starts daemons
func StartDaemons() {
	if config.ConfigIni["daemons"] == "null" {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	utils.CancelFunc = cancel
	utils.ReturnCh = make(chan string)

	daemonsToStart := serverList
	if utils.Mobile() {
		daemonsToStart = mobileList
	} else if *utils.TestRollBack == 1 {
		daemonsToStart = rollbackList
	}

	if len(config.ConfigIni["daemons"]) > 0 {
		daemonsToStart = strings.Split(config.ConfigIni["daemons"], ",")
	}

	for _, name := range daemonsToStart {
		handler, ok := daemonsList[name]
		if ok {
			go daemonLoop(ctx, name, handler, utils.ReturnCh)
			utils.DaemonsCount++
			continue
		}

		logger.Errorf("unknown daemon name: %s", name)

	}
}

func GetHostPort(h string) string {
	if strings.Contains(h, ":") {
		return h
	}
	return h + ":" + consts.TCP_PORT
}
