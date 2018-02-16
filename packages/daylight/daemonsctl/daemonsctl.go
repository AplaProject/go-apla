package daemonsctl

import (
	conf "github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/daemons"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/tcpserver"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	"strings"

	log "github.com/sirupsen/logrus"
)

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

// RunAllDaemons start daemons, load contracts and tcpserver
func RunAllDaemons() error {
	err := syspar.SysUpdate(nil)
	if err != nil {
		log.Errorf("can't read system parameters: %s", utils.ErrInfo(err))
		return err
	}

	log.Info("load contracts")
	if err := smart.LoadContracts(nil); err != nil {
		log.Errorf("Load Contracts error: %s", err)
		return err
	}

	log.Info("start daemons")

	if conf.Config.StartDaemons != "null" {
		daemonsToStart := serverList
		if len(conf.Config.StartDaemons) > 0 {
			daemonsToStartParam := strings.Split(conf.Config.StartDaemons, ",")
			daemonsToStart = daemonsToStartParam
		} else if *conf.TestRollBack {
			daemonsToStart = rollbackList
		}

		daemons.StartDaemons(daemonsToStart)
	}

	err = tcpserver.TcpListener(conf.Config.TCPServer.Str())
	if err != nil {
		log.Errorf("can't start tcp servers, stop")
		return err
	}

	return nil
}

func RunSpecificDaemons(daemonsToStart []string) {
	daemons.StartDaemons(daemonsToStart)
}
