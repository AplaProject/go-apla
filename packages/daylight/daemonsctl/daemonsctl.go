package daemonsctl

import (
	conf "github.com/GenesisCommunity/go-genesis/packages/conf"
	"github.com/GenesisCommunity/go-genesis/packages/config/syspar"
	"github.com/GenesisCommunity/go-genesis/packages/daemons"
	"github.com/GenesisCommunity/go-genesis/packages/smart"
	"github.com/GenesisCommunity/go-genesis/packages/tcpserver"
	"github.com/GenesisCommunity/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// RunAllDaemons start daemons, load contracts and tcpserver
func RunAllDaemons() error {
	if !conf.Config.IsSupportingVDE() {
		logEntry := log.WithFields(log.Fields{"daemon_name": "block_collection"})

		daemons.InitialLoad(logEntry)
		err := syspar.SysUpdate(nil)
		if err != nil {
			log.Errorf("can't read system parameters: %s", utils.ErrInfo(err))
			return err
		}

		if data, ok := block.GetDataFromFirstBlock(); ok {
			syspar.SetFirstBlockData(data)
		}
	}

	log.Info("load contracts")
	if err := smart.LoadContracts(nil); err != nil {
		log.Errorf("Load Contracts error: %s", err)
		return err
	}

	log.Info("start daemons")
	daemons.StartDaemons()

	if err := tcpserver.TcpListener(conf.Config.TCPServer.Str()); err != nil {
		log.Errorf("can't start tcp servers, stop")
		return err
	}

	return nil
}
