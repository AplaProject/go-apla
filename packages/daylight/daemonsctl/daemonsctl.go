package daemonsctl

import (
	conf "github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/daemons"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/tcpserver"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

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
	daemons.StartDaemons()

	if *conf.IsVDEMasterMode || *conf.IsVDEMode {
		return nil
	}

	err = tcpserver.TcpListener(conf.Config.TCPServer.Str())
	if err != nil {
		log.Errorf("can't start tcp servers, stop")
		return err
	}

	return nil
}
