package modes

import (
	"github.com/GenesisKernel/go-genesis/packages/api"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/julienschmidt/httprouter"
)

func InitVDEMode(config *conf.VDEConfig) *VDE {
	mode := &VDE{
		VDEConfig: config,
		api:       api.CreateDefaultRouter(),
	}

	return mode
}

type VDE struct {
	*conf.VDEConfig
	api *httprouter.Router
}

func (mode *VDE) Start(exitFunc func(int), gormInit func(conf.DBConfig), listenerFunc func(string, *httprouter.Router)) {
	gormInit(mode.DB)

	listenerFunc(mode.VDEConfig.HTTP.Str(), mode.api)
}

func (mode *VDE) DaemonList() []string {
	return []string{
		"Notificator",
		"Scheduler",
	}
}
