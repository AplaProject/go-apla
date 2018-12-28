package modes

import (
	"context"
	"time"

	"github.com/AplaProject/go-apla/packages/api"
	"github.com/AplaProject/go-apla/packages/block"
	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/daemons"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/network/tcpserver"
	"github.com/AplaProject/go-apla/packages/service"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/AplaProject/go-apla/packages/utils"
	log "github.com/sirupsen/logrus"
)

type BCEcosysLookupGetter struct{}

func (g BCEcosysLookupGetter) GetEcosystemLookup() ([]int64, []string, error) {
	return model.GetAllSystemStatesIDs()
}

type OBSEcosystemLookupGetter struct{}

func (g OBSEcosystemLookupGetter) GetEcosystemLookup() ([]int64, []string, error) {
	return []int64{1}, []string{"Platform ecosystem"}, nil
}

func BuildEcosystemLookupGetter() types.EcosystemLookupGetter {
	if conf.Config.IsSupportingOBS() {
		return OBSEcosystemLookupGetter{}
	}

	return BCEcosysLookupGetter{}
}

type BCEcosysIDValidator struct {
	logger *log.Entry
}

func (v BCEcosysIDValidator) Validate(id, clientID int64, le *log.Entry) (int64, error) {
	if clientID <= 0 {
		return clientID, nil
	}

	count, err := model.GetNextID(nil, "1_ecosystems")
	if err != nil {
		le.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id of ecosystems")
		return 0, err
	}

	if clientID >= count {
		le.WithFields(log.Fields{"state_id": clientID, "count": count, "type": consts.ParameterExceeded}).Error("ecosystem is larger then max count")
		return 0, api.ErrEcosystemNotFound
	}

	return clientID, nil
}

type OBSEcosysIDValidator struct{}

func (OBSEcosysIDValidator) Validate(id, clientID int64, le *log.Entry) (int64, error) {
	return consts.DefaultOBS, nil
}

func GetEcosystemIDValidator() types.EcosystemIDValidator {
	if conf.Config.IsSupportingOBS() {
		return OBSEcosysIDValidator{}
	}

	return BCEcosysIDValidator{}
}

type BCEcosystemNameGetter struct{}

func (ng BCEcosystemNameGetter) GetEcosystemName(id int64) (string, error) {
	ecosystem := &model.Ecosystem{}
	found, err := ecosystem.Get(id)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting ecosystem from db")
		return "", err
	}

	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "id": id, "error": api.ErrEcosystemNotFound}).Error("ecosystem not found")
		return "", err
	}

	return ecosystem.Name, nil
}

type OBSEcosystemNameGetter struct{}

func (ng OBSEcosystemNameGetter) GetEcosystemName(id int64) (string, error) {
	return "Platform ecosystem", nil
}

func BuildEcosystemNameGetter() types.EcosystemNameGetter {
	if conf.Config.IsSupportingOBS() {
		return OBSEcosystemNameGetter{}
	}

	return BCEcosystemNameGetter{}
}

// BCDaemonLoader allow load blockchain daemons
type BCDaemonLoader struct {
	logger            *log.Entry
	DaemonListFactory types.DaemonListFactory
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

	mode := "Public blockchain"
	if syspar.IsPrivateBlockchain() {
		mode = "Private Blockchain"
	}

	logMode(l.logger, mode)

	l.logger.Info("load contracts")
	if err := smart.LoadContracts(); err != nil {
		log.Errorf("Load Contracts error: %s", err)
		return err
	}

	l.logger.Info("start daemons")
	daemons.StartDaemons(ctx, l.DaemonListFactory.GetDaemonsList())

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
	logger            *log.Entry
	DaemonListFactory types.DaemonListFactory
}

// Load loads obs daemons
func (l OBSDaemonLoader) Load(ctx context.Context) error {

	if err := syspar.SysUpdate(nil); err != nil {
		l.logger.Errorf("can't read system parameters: %s", utils.ErrInfo(err))
		return err
	}

	logMode(l.logger, conf.Config.OBSMode)
	l.logger.Info("load contracts")
	if err := smart.LoadContracts(); err != nil {
		l.logger.Errorf("Load Contracts error: %s", err)
		return err
	}

	l.logger.Info("start daemons")
	daemons.StartDaemons(ctx, l.DaemonListFactory.GetDaemonsList())

	return nil
}

func GetDaemonLoader() types.DaemonLoader {
	if conf.Config.IsSupportingOBS() {
		return OBSDaemonLoader{
			logger:            log.WithFields(log.Fields{"loader": "obs_daemon_loader"}),
			DaemonListFactory: OBSDaemonsListFactory{},
		}
	}

	return BCDaemonLoader{
		logger:            log.WithFields(log.Fields{"loader": "blockchain_daemon_loader"}),
		DaemonListFactory: BlockchainDaemonsListsFactory{},
	}
}

func logMode(logger *log.Entry, mode string) {
	logLevel := log.GetLevel()
	log.SetLevel(log.InfoLevel)
	logger.WithFields(log.Fields{"mode": mode}).Info("Node running mode")
	log.SetLevel(logLevel)
}
