package api

import (
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/publisher"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func configOptionHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	if len(params["option"]) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject, "error": "option not specified"}).Error("on getting option in config handler")
		errorResponse(w, newError(errServer, http.StatusBadRequest))
		return
	}

	switch params["option"] {
	case "centrifugo":
		centrifugoAddressHandler(w, r)
		return
	}

	errorResponse(w, newError(errServer, http.StatusBadRequest))
}

func replaceHttpSchemeToWs(centrifugoURL string) string {
	if strings.HasPrefix(centrifugoURL, "http:") {
		return strings.Replace(centrifugoURL, "http:", "ws:", -1)
	} else if strings.HasPrefix(centrifugoURL, "https:") {
		return strings.Replace(centrifugoURL, "https:", "wss:", -1)
	}
	return centrifugoURL
}

func centrifugoAddressHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	if _, err := publisher.GetStats(); err != nil {
		logger.WithFields(log.Fields{"type": consts.CentrifugoError, "error": err}).Warn("on getting centrifugo stats")
		errorResponse(w, err)
		return
	}

	jsonResponse(w, replaceHttpSchemeToWs(conf.Config.Centrifugo.URL))
}
