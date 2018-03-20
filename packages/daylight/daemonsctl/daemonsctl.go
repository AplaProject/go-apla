package daemonsctl

import (
	conf "github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/daemons"
	"github.com/GenesisKernel/go-genesis/packages/parser"
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

	// If list of nodes is empty, then used node from the first block
	if syspar.GetNumberOfNodes() == 0 {
		if keyID, publicKey, ok := parser.GetKeysFromFirstBlock(); ok {
			syspar.AddFullNodeKeys(keyID, publicKey)
		}
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
