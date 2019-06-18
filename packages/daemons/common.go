// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package daemons

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/statsd"
	"github.com/AplaProject/go-apla/packages/utils"

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
	"Scheduler":         Scheduler,
	"ExternalNetwork":   ExternalNetwork,
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
		sleepTime:     100 * time.Millisecond,
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
func StartDaemons(ctx context.Context, daemonsToStart []string) {
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

	// ctx, cancel := context.WithCancel(context.Background())
	// utils.CancelFunc = cancel
	// utils.ReturnCh = make(chan string)

	if conf.Config.TestRollBack {
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
