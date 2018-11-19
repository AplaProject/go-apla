package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type contractsResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

func getContractsHandler(w http.ResponseWriter, r *http.Request) {
	form := &paginatorForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	client := getClient(r)
	logger := getLogger(r)

	contract := &model.Contract{}
	contract.EcosystemID = client.EcosystemID

	count, err := contract.CountByEcosystem()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting table records count")
		errorResponse(w, err)
		return
	}

	contracts, err := contract.GetListByEcosystem(form.Offset, form.Limit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all")
		errorResponse(w, err)
		return
	}

	list := make([]map[string]string, len(contracts))
	for i, c := range contracts {
		list[i] = c.ToMap()
		list[i]["address"] = converter.AddressToString(c.WalletID)
	}

	if len(list) == 0 {
		list = nil
	}

	jsonResponse(w, &listResult{
		Count: converter.Int64ToStr(count),
		List:  list,
	})
}
