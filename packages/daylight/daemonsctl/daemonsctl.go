package daemonsctl

import (
	"context"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/block"
	conf "github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/daemons"
	"github.com/GenesisKernel/go-genesis/packages/network/tcpserver"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// RunAllDaemons start daemons, load contracts and tcpserver
func RunAllDaemons(ctx context.Context) error {
	loader := GetDaemonLoader(conf.Config.IsSupportingOBS())

	return loader.Load(ctx)
}

// DaemonLoader allow implement different ways for loading daemons
type DaemonLoader interface {
	Load(context.Context) error
}

// BCDaemonLoader allow load blockchain daemons
type BCDaemonLoader struct {
	logger *log.Entry
}

// Load loads blockchain daemons
func (l BCDaemonLoader) Load(ctx context.Context) error {

	daemons.InitialLoad(l.logger)

	if err := syspar.SysUpdate(nil); err != nil {
		log.Errorf("can't read system parameters: %s", utils.ErrInfo(err))
		return err
	}

	if data, ok := block.GetDataFromFirstBlock(); ok {
		syspar.SetFirstBlockData(data)
	}

	l.logger.Info("load contracts")
	if err := smart.LoadContracts(); err != nil {
		log.Errorf("Load Contracts error: %s", err)
		return err
	}

	l.logger.Info("start daemons")
	daemons.StartDaemons(ctx)

	if err := tcpserver.TcpListener(conf.Config.TCPServer.Str()); err != nil {
		log.Errorf("can't start tcp servers, stop")
		return err
	}

	var availableBCGap int64 = consts.AvailableBCGap
	if syspar.GetRbBlocks1() > consts.AvailableBCGap {
		availableBCGap = syspar.GetRbBlocks1() - consts.AvailableBCGap
	}

	blockGenerationDuration := time.Millisecond * time.Duration(syspar.GetMaxBlockGenerationTime())
	blocksGapDuration := time.Second * time.Duration(syspar.GetGapsBetweenBlocks())
	blockGenerationTime := blockGenerationDuration + blocksGapDuration

	checkingInterval := blockGenerationTime * time.Duration(syspar.GetRbBlocks1()-consts.DefaultNodesConnectDelay)
	na := service.NewNodeRelevanceService(availableBCGap, checkingInterval)
	na.Run(ctx)

	if err := service.InitNodesBanService(); err != nil {
		l.logger.WithError(err).Fatal("Can't init ban service")
	}

	return nil
}

// OBSDaemonLoader allows load obs daemons
type OBSDaemonLoader struct {
	logger *log.Entry
}

// Load loads obs daemons
func (l OBSDaemonLoader) Load(ctx context.Context) error {

	daemons.InitialLoad(l.logger)

	if err := syspar.SysUpdate(nil); err != nil {
		log.Errorf("can't read system parameters: %s", utils.ErrInfo(err))
		return err
	}

	l.logger.Info("load contracts")
	if err := smart.LoadContracts(); err != nil {
		log.Errorf("Load Contracts error: %s", err)
		return err
	}

	l.logger.Info("start daemons")
	daemons.StartDaemons(ctx)

	return nil
}

func GetDaemonLoader(obs bool) DaemonLoader {
	if obs {
		return OBSDaemonLoader{
			logger: log.WithFields(log.Fields{"loader": "obs_daemon_loader"}),
		}
	}

	return BCDaemonLoader{
		logger: log.WithFields(log.Fields{"loader": "blockchain_daemon_loader"}),
	}
}
