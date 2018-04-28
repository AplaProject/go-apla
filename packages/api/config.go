package api

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/GenesisKernel/go-genesis/packages/conf"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/publisher"
	log "github.com/sirupsen/logrus"
)

func configOptionHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	if len(params["option"]) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject, "error": "option not specified"}).Error("on getting option in config handler")
		errorResponse(w, errServer, http.StatusBadRequest)
		return
	}

	switch params["option"] {
	case "centrifugo":
		centrifugoAddressHandler(w, r)
		return
	}

	errorResponse(w, errServer, http.StatusBadRequest)
}

func centrifugoAddressHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	if _, err := publisher.GetStats(); err != nil {
		logger.WithFields(log.Fields{"type": consts.CentrifugoError, "error": err}).Warn("on getting centrifugo stats")
		errorResponse(w, err, http.StatusNotFound)
		return
	}

	jsonResponse(w, conf.Config.Centrifugo.URL)
}
