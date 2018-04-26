package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/conf"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/publisher"
	log "github.com/sirupsen/logrus"
)

type configOptionHandler func(w http.ResponseWriter, option string) error

func getConfigOption(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	option := data.params["option"].(string)
	if len(option) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject, "error": "option not specified"}).Error("on getting option in config handler")
		return errorAPI(w, "E_SERVER", http.StatusBadRequest)
	}

	var err error
	switch option {
	case "centrifugo":
		err = centrifugoAddressHandler(w, data)
		break
	default:
		return errorAPI(w, "E_SERVER", http.StatusBadRequest)
	}

	return err
}

func centrifugoAddressHandler(w http.ResponseWriter, data *apiData) error {
	if _, err := publisher.GetStats(); err != nil {
		log.WithFields(log.Fields{"type": consts.CentrifugoError, "error": err}).Warn("on getting centrifugo stats")
		return errorAPI(w, err, http.StatusNotFound)
	}

	data.result = conf.Config.Centrifugo.URL
	return nil
}
