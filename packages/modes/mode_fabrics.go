package modes

import (
	"context"
	"errors"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/block"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/daemons"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/network/tcpserver"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	log "github.com/sirupsen/logrus"
)

var ErrEcosystemNotFound = errors.New("Ecosystem not found")

type EcosystemLookupGetter interface {
	GetEcosystemLookup() ([]int64, []string, error)
}

type BCEcosysLookupGetter struct{}

func (g BCEcosysLookupGetter) GetEcosystemLookup() ([]int64, []string, error) {
	return model.GetAllSystemStatesIDs()
}

type OBSEcosystemLookupGetter struct{}

func (g OBSEcosystemLookupGetter) GetEcosystemLookup() ([]int64, []string, error) {
	return []int64{1}, []string{"Platform ecosystem"}, nil
}

func BuildEcosystemLookupGetter() EcosystemLookupGetter {
	if conf.Config.IsSupportingOBS() {
		return OBSEcosystemLookupGetter{}
	}

	return BCEcosysLookupGetter{}
}

func ValidateEcosysID(formID, clientID int64, logger *log.Entry) (int64, error) {
	if conf.Config.IsSupportingOBS() {
		return consts.DefaultOBS, nil
	}

	if formID <= 0 {
		return clientID, nil
	}

	count, err := model.GetNextID(nil, "1_ecosystems")
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id of ecosystems")
		return 0, err
	}

	if formID >= count {
		logger.WithFields(log.Fields{"state_id": formID, "count": count, "type": consts.ParameterExceeded}).Error("ecosystem is larger then max count")
		return 0, ErrEcosystemNotFound
	}

	return formID, nil
}

func GetEcosystemName(id int64) (string, error) {
	if conf.Config.IsSupportingOBS() {
		return "Platform ecosystem", nil
	}

	ecosystem := &model.Ecosystem{}
	found, err := ecosystem.Get(id)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting ecosystem from db")
		return "", err
	}

	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "id": id, "error": ErrEcosystemNotFound}).Error("ecosystem not found")
		return "", err
	}

	return ecosystem.Name, nil

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
	daemons.StartDaemons(ctx, GetDaemonsToStart())

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
	daemons.StartDaemons(ctx, GetDaemonsToStart())

	return nil
}

func GetDaemonLoader() DaemonLoader {
	if conf.Config.IsSupportingOBS() {
		return OBSDaemonLoader{
			logger: log.WithFields(log.Fields{"loader": "obs_daemon_loader"}),
		}
	}

	return BCDaemonLoader{
		logger: log.WithFields(log.Fields{"loader": "blockchain_daemon_loader"}),
	}
}
