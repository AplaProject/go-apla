package daemonsctl

import (
	"context"

	"github.com/GenesisKernel/go-genesis/packages/block"
	conf "github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/daemons"
	"github.com/GenesisKernel/go-genesis/packages/network/tcpserver"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// RunAllDaemons start daemons, load contracts and tcpserver
func RunAllDaemons(ctx context.Context) error {
	err := syspar.SysUpdate(nil)
	if err != nil {
		log.Errorf("can't read system parameters: %s", utils.ErrInfo(err))
		return err
	}

	if !conf.Config.IsSupportingVDE() {
		logEntry := log.WithFields(log.Fields{"daemon_name": "block_collection"})

		daemons.InitialLoad(logEntry)

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
	daemons.StartDaemons(ctx)

	if err := tcpserver.TcpListener(conf.Config.TCPServer.Str()); err != nil {
		log.Errorf("can't start tcp servers, stop")
		return err
	}

	return nil
}
