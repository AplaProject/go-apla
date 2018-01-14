package main

import (
	"flag"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/tools/desync_monitor/config"
	"github.com/AplaProject/go-apla/tools/desync_monitor/query"

	log "github.com/sirupsen/logrus"
)

const confPathFlagName = "confPath"
const nodesListFlagName = "nodesList"
const daemonModeFlagName = "daemonMode"
const queryingPeriodFlagName = "queryingPeriod"
const alertEmailFlagName = "alertEmail"
const alertMessageFlagName = "alertMessage"

var configPath *string = flag.String(confPathFlagName, "config.toml", "path to desync monitor config")
var nodesList *string = flag.String(nodesListFlagName, "127.0.0.1:7079", "which nodes to query, in format url1,url2,url3")
var daemonMode *bool = flag.Bool(daemonModeFlagName, false, "start as daemon")
var queryingPeriod *int = flag.Int(queryingPeriodFlagName, 1, "period of querying nodes in seconds, if started as daemon")
var alertEmail *string = flag.String(alertEmailFlagName, "alert@apla.io", "email adress to send alert")
var alertMessage *string = flag.String(alertMessageFlagName, "nodes unynced!!!", "alert message to send")

func minElement(slice []int64) int64 {
	var min int64 = math.MaxInt64
	for _, blockID := range slice {
		if blockID < min {
			min = blockID
		}
	}
	return min
}

func flagsOverrideConfig(conf *config.Config) {
	flag.Visit(func(flag *flag.Flag) {
		switch flag.Name {
		case nodesListFlagName:
			nodesList := strings.Split(*nodesList, ",")
			conf.NodesList = nodesList
		case daemonModeFlagName:
			conf.Daemon.DaemonMode = *daemonMode
		case queryingPeriodFlagName:
			conf.Daemon.QueryingPeriod = *queryingPeriod
		case alertEmailFlagName:
			conf.Alert.Email = *alertEmail
		case alertMessageFlagName:
			conf.Alert.Message = *alertMessage
		}
	})
}

func sendEmail(emailAddress, message string) {
	fmt.Println("SENT ", message, " TO ADDRESS ", emailAddress)
}

func monitor(conf *config.Config) {
	maxBlockIDs, err := query.MaxBlockIDs(conf.NodesList)
	if err != nil {
		sendEmail(conf.Alert.Email, "problem getting node max block id :"+err.Error())
		return
	}
	blockInfos, err := query.BlockInfo(conf.NodesList, minElement(maxBlockIDs))
	if err != nil {
		sendEmail(conf.Alert.Email, "problem getting node block info :"+err.Error())
		return
	}
	hash2Node := map[string][]string{}
	for node, blockInfo := range blockInfos {
		rollbacksHash := string(blockInfo.RollbacksHash)
		if _, ok := hash2Node[rollbacksHash]; !ok {
			hash2Node[rollbacksHash] = []string{}
		}
		hash2Node[rollbacksHash] = append(hash2Node[rollbacksHash], node)
	}
	if len(hash2Node) > 1 {
		sendEmail(conf.Alert.Email, "nodes unsynced!!!")
	}
}

func main() {
	flag.Parse()
	conf := &config.Config{}
	if err := conf.Read(*configPath); err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("reading config")
	}
	flagsOverrideConfig(conf)
	if conf.Daemon.DaemonMode {
		ticker := time.NewTicker(time.Second * time.Duration(conf.Daemon.QueryingPeriod))
		for _ = range ticker.C {
			monitor(conf)
		}
	} else {
		monitor(conf)
	}
}
