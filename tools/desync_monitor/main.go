package main

import (
	"flag"
	"fmt"
	"math"
	"net/smtp"
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
const alertMessageToFlagName = "alertMessageTo"
const alertMessageSubjFlagName = "alertMessageSubj"
const alertMessageFromFlagName = "alertMessageFrom"
const smtpHostFlagName = "smtpHost"
const smtpPortFlagName = "smtpPort"
const smtpUsernameFlagName = "smtpUsername"
const smtpPasswordFlagName = "smtpPassword"

var configPath *string = flag.String(confPathFlagName, "config.toml", "path to desync monitor config")
var nodesList *string = flag.String(nodesListFlagName, "127.0.0.1:7079", "which nodes to query, in format url1,url2,url3")
var daemonMode *bool = flag.Bool(daemonModeFlagName, false, "start as daemon")
var queryingPeriod *int = flag.Int(queryingPeriodFlagName, 1, "period of querying nodes in seconds, if started as daemon")

var alertMessageTo *string = flag.String(alertMessageToFlagName, "alert@apla.io", "email adress to send alert")
var alertMessageSubj *string = flag.String(alertMessageSubjFlagName, "problem with nodes synchronization", "alert message subject")
var alertMessageFrom *string = flag.String(alertMessageFromFlagName, "monitor@apla.io", "email adress from witch to send alert")

var smtpHost *string = flag.String(smtpHostFlagName, "", "host of smtp server, to send alert email")
var smtpPort *int = flag.Int(smtpPortFlagName, 25, "port of smtp server")
var smtpUsername *string = flag.String(smtpUsernameFlagName, "", "login to smtp server")
var smtpPassword *string = flag.String(smtpPasswordFlagName, "", "password to smtp server")

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
		case alertMessageToFlagName:
			conf.AlertMessage.To = *alertMessageTo
		case alertMessageSubjFlagName:
			conf.AlertMessage.Subject = *alertMessageSubj
		case alertMessageFromFlagName:
			conf.AlertMessage.From = *alertMessageFrom
		case smtpHostFlagName:
			conf.Smtp.Host = *smtpHost
		case smtpPortFlagName:
			conf.Smtp.Port = *smtpPort
		case smtpUsernameFlagName:
			conf.Smtp.Username = *smtpUsername
		case smtpPasswordFlagName:
			conf.Smtp.Password = *smtpPassword
		}
	})
}

func sendEmail(smtpConf *config.Smtp, alertConf *config.AlertMessage, message string) error {
	auth := smtp.PlainAuth("", smtpConf.Username, smtpConf.Password, smtpConf.Host)
	to := []string{alertConf.To}
	msg := []byte(fmt.Sprintf("From: %s\r\n", alertConf.From) +
		fmt.Sprintf("To: %s\r\n", alertConf.To) +
		fmt.Sprintf("Subject: %s\r\n", alertConf.Subject) +
		"\r\n" +
		fmt.Sprintf("%s\r\n", message))
	err := smtp.SendMail(fmt.Sprintf("%s:%d", smtpConf.Host, smtpConf.Port), auth, alertConf.From, to, msg)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("sending email")
	}
	return err
}

func monitor(conf *config.Config) {
	maxBlockIDs, err := query.MaxBlockIDs(conf.NodesList)
	if err != nil {
		sendEmail(&conf.Smtp, &conf.AlertMessage, "problem getting node max block id :"+err.Error())
		return
	}
	blockInfos, err := query.BlockInfo(conf.NodesList, minElement(maxBlockIDs))
	if err != nil {
		sendEmail(&conf.Smtp, &conf.AlertMessage, "problem getting node block info :"+err.Error())
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
		hash2NodeStrResults := []string{}
		for k, v := range hash2Node {
			hash2NodeStrResults = append(hash2NodeStrResults, fmt.Sprintf("%x: %s", k, v))
		}
		sendEmail(&conf.Smtp, &conf.AlertMessage, fmt.Sprintf("nodes unsynced. Rollback hashes are: %s", strings.Join(hash2NodeStrResults, ",")))
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
