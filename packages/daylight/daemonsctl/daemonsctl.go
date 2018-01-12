package daemonsctl

import (
	conf "github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/daemons"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/tcpserver"
	"github.com/AplaProject/go-apla/packages/utils"
	logging "github.com/op/go-logging"
)

var (
	log    = logging.MustGetLogger("daylight")
	format = logging.MustStringFormatter("%{time:15:04:05.000} %{shortfile} %{shortfunc} [%{level:.4s}] %{message}")
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

	err = tcpserver.TcpListener(conf.Config.TCPServer.Str())
	if err != nil {
		log.Errorf("can't start tcp servers, stop")
		return err
	}

	return nil
}
