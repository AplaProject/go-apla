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
package daemons

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GenesisCommunity/go-genesis/packages/conf"
	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/statsd"
	"github.com/GenesisCommunity/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

var (
	// MonitorDaemonCh is monitor daemon channel
	MonitorDaemonCh = make(chan []string, 100)
)

type daemon struct {
	goRoutineName string
	sleepTime     time.Duration
	logger        *log.Entry
}

var daemonsList = map[string]func(context.Context, *daemon) error{
	"BlocksCollection":  BlocksCollection,
	"BlockGenerator":    BlockGenerator,
	"Disseminator":      Disseminator,
	"QueueParserTx":     QueueParserTx,
	"QueueParserBlocks": QueueParserBlocks,
	"Confirmations":     Confirmations,
	"Notificator":       Notificate,
	"Scheduler":         Scheduler,
}

var serverList = []string{
	"BlocksCollection",
	"BlockGenerator",
	"QueueParserTx",
	"QueueParserBlocks",
	"Disseminator",
	"Confirmations",
	"Notificator",
	"Scheduler",
}

var rollbackList = []string{
	"BlocksCollection",
	"Confirmations",
}

func daemonLoop(ctx context.Context, goRoutineName string, handler func(context.Context, *daemon) error, retCh chan string) {
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
	handler(ctx, d)
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
			handler(ctx, d)
			statsd.Client.TimingDuration(counterName+statsd.Time, time.Now().Sub(startTime), 1.0)
		}
	}
}

// StartDaemons starts daemons
func StartDaemons() {
	if conf.Config.StartDaemons == "null" {
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
	if len(conf.Config.StartDaemons) > 0 {
		daemonsToStart = strings.Split(conf.Config.StartDaemons, ",")
	} else if *conf.TestRollBack {
		daemonsToStart = rollbackList
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
	return fmt.Sprintf("%s:%d", h, consts.DEFAULT_TCP_PORT)
}
