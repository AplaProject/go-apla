package modes

import (
	"net/http"

	"github.com/AplaProject/go-apla/packages/api"
	"github.com/AplaProject/go-apla/packages/conf"
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

	return r.GetAPI()
}
