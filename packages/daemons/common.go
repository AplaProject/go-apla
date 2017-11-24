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

	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/statsd"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

var (
	MonitorDaemonCh = make(chan []string, 100)
)

type daemon struct {
	goRoutineName string
	sleepTime     time.Duration
	logger        *log.Entry
}

func init() {
	flag.Parse()
}

var daemonsList = map[string]func(*daemon, context.Context) error{
	"BlocksCollection":   BlocksCollection,
	"BlockGenerator":     BlockGenerator,
	"CreatingBlockchain": CreatingBlockchain,
	"Disseminator":       Disseminator,
	"QueueParserTx":      QueueParserTx,
	"QueueParserBlocks":  QueueParserBlocks,
	"Confirmations":      Confirmations,
	"Notificator":        Notificate,
}

var serverList = []string{
	"BlocksCollection",
	"CreatingBlockchain",
	"BlockGenerator",
	"QueueParserTx",
	"QueueParserBlocks",
	"Disseminator",
	"Confirmations",
	"Notificator",
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
	logger := log.WithFields(log.Fields{"daemon_name": goRoutineName})
	defer func() {
		if r := recover(); r != nil {
			logger.WithFields(log.Fields{"type": consts.PanicRecoveredError, "error": r}).Error("panic in daemon")
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
		logger:        logger,
	}

	startTime := time.Now()
	counterName := statsd.DaemonCounterName(goRoutineName)
	handler(d, ctx)
	statsd.Client.TimingDuration(counterName+statsd.Time, time.Now().Sub(startTime), 1.0)

	for {
		select {
		case <-ctx.Done():
			logger.Info("daemon done his work")
			retCh <- goRoutineName
			return

		case <-time.After(d.sleepTime):
			MonitorDaemonCh <- []string{d.goRoutineName, converter.Int64ToStr(time.Now().Unix())}
			startTime := time.Now()
			counterName := statsd.DaemonCounterName(goRoutineName)
			handler(d, ctx)
			statsd.Client.TimingDuration(counterName+statsd.Time, time.Now().Sub(startTime), 1.0)
		}
	}
}

// StartDaemons starts daemons
func StartDaemons() {
	if config.ConfigIni["daemons"] == "null" {
		return
	}

	go WaitStopTime()

	daemonsTable := make(map[string]string)
	go func() {
		for {
			daemonNameAndTime := <-MonitorDaemonCh
			daemonsTable[daemonNameAndTime[0]] = daemonNameAndTime[1]
			if time.Now().Unix()%10 == 0 {
				log.Debug("daemonsTable: %v\n", daemonsTable)
			}
		}
	}()

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
	log.WithFields(log.Fields{"daemons_to_start": daemonsToStart}).Info("starting daemons")

	for _, name := range daemonsToStart {
		handler, ok := daemonsList[name]
		if ok {
			go daemonLoop(ctx, name, handler, utils.ReturnCh)
			log.WithFields(log.Fields{"daemon_name": name}).Info("started")
			utils.DaemonsCount++
			continue
		}

		log.WithFields(log.Fields{"daemon_name": name}).Warning("unknown daemon name")
	}
}

func getHostPort(h string) string {
	if strings.Contains(h, ":") {
		return h
	}
	return h + ":" + consts.TCP_PORT
}
