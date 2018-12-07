package modes

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/api"
	"github.com/GenesisKernel/go-genesis/packages/conf"
)

func RegisterRoutes() http.Handler {
	m := api.Mode{
		EcosysIDValidator:  GetEcosystemIDValidator(),
		EcosysNameGetter:   BuildEcosystemNameGetter(),
		EcosysLookupGetter: BuildEcosystemLookupGetter(),
		ContractRunner:     GetSmartContractRunner(),
		ClientTxProcessor:  GetClientTxPreprocessor(),
	}

	r := api.NewRouter(m)
	if !conf.Config.IsSupportingOBS() {
		m.SetBlockchainRoutes(r)
	}
	return r
}
