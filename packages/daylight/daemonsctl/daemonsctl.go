package daemonsctl

import (
	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
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

// RunAllDaemons is starting all daemons
func RunAllDaemons() error {
	err := syspar.SysUpdate()
	if err != nil {
		log.Errorf("can't read system parameters: %s", utils.ErrInfo(err))
		return err
	}

	log.Info("start daemons")
	daemons.StartDaemons()

	if err := smart.LoadContracts(nil); err != nil {
		log.Errorf("Load Contracts error: %s", err)
		return err
	}

	err = tcpserver.TcpListener(*utils.TCPHost + ":" + consts.TCP_PORT)
	if err != nil {
		log.Errorf("can't start tcp servers, stop")
		return err
	}

	return nil
}
