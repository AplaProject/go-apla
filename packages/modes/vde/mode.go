package vde

import (
	"github.com/GenesisKernel/go-genesis/packages/api"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/daemons"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/modes"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// Config config for VDE mode
type Config struct {
	DB          conf.DBConfig
	HTTP        conf.HostPort
	Centrifugo  conf.CentrifugoConfig
	Autoupdate  conf.AutoupdateConfig
	WorkDir     string
	LogLevel    string
	LogFileName string
}

// VDE represent VDE mode implement NodeMode interface
type VDE struct {
	*Config
	api        *httprouter.Router
	daemonList []string
}

// Init init VDE Mode
func Init(config *Config) *VDE {
	mode := &VDE{
		Config: config,
		api:    api.CreateDefaultRouter(true),
		daemonList: []string{
			"Notificator",
			"Scheduler",
		},
	}

	return mode
}

// Start implement Start func of NodeMode interface
func (mode *VDE) Start(exitFunc func(int), gormInit func(conf.DBConfig)) {
	gormInit(mode.DB)

	if model.DBConn != nil {
		// The installation process is already finished (where user has specified DB and where wallet has been restarted)
		err := syspar.SysUpdate(nil)
		if err != nil {
			log.Errorf("can't read system parameters: %s", utils.ErrInfo(err))
			exitFunc(1)
		}

		log.Info("load contracts")
		if err := smart.LoadContracts(nil); err != nil {
			log.Errorf("Load Contracts error: %s", err)
			exitFunc(1)
		}

		log.Info("start daemons")
		daemons.StartDaemons(mode.daemonList)

		log.Info("Daemons started")
	}
}

// DaemonList implement func of NodeMode interface
func (mode *VDE) DaemonList() []string {
	return mode.daemonList
}

// Stop implement func of NodeMode interface
func (mode *VDE) Stop() {
	log.Infoln("VDE mode stopped")
}

// API returns mode api
func (mode *VDE) API() *httprouter.Router {
	return mode.api
}

// Mode returns node type
func (mode *VDE) Mode() modes.ModeType {
	return modes.TypeVDE
}
