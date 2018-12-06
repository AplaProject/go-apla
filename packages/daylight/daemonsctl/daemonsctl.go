package daemonsctl

import (
	"context"

	"github.com/AplaProject/go-apla/packages/blockchain"
	conf "github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/daemons"
	"github.com/AplaProject/go-apla/packages/network/tcpserver"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

// RunAllDaemons start daemons, load contracts and tcpserver
func RunAllDaemons(ctx context.Context) error {
	if !conf.Config.IsSupportingVDE() {
		logEntry := log.WithFields(log.Fields{"daemon_name": "block_collection"})

		daemons.InitialLoad(logEntry)

		err := syspar.SysUpdate(nil)
		if err != nil {
			log.Errorf("can't read system parameters: %s", utils.ErrInfo(err))
			return err
		}
		syspar.SetFirstBlockData()
	} else {
		err := syspar.SysUpdate(nil)
		if err != nil {
			log.Errorf("can't read system parameters: %s", utils.ErrInfo(err))
			return err
		}

	}

	_, _, found, err := blockchain.GetLastBlock(nil)
	if err != nil {
		log.WithError(err).Error("Getting first block")
		return err
	}
	if found {
		log.Info("load contracts")
		if err := smart.LoadContracts(); err != nil {
			log.Errorf("Load Contracts error: %s", err)
			return err
		}
	}

	log.Info("start daemons")
	daemons.StartDaemons(ctx)

	if err := tcpserver.TcpListener(conf.Config.TCPServer.Str()); err != nil {
		log.Errorf("can't start tcp servers, stop")
		return err
	}

	return nil
}
