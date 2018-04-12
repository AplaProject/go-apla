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
		err = centrifugoAddressHandler(w)
	default:
		err = errorAPI(w, "E_SERVER", http.StatusBadRequest)
	}

	return err
}

func centrifugoAddressHandler(w http.ResponseWriter) error {
	if _, err := publisher.GetStats(); err != nil {
		log.WithFields(log.Fields{"type": consts.CentrifugoError, "error": err}).Warn("on getting centrifugo stats")
		return errorAPI(w, "E_SERVER", http.StatusNotFound)
	}

	if _, err := w.Write([]byte(conf.Config.Centrifugo.URL)); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on write centrifugo address response")
		return err
	}
	return nil
}
