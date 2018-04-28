package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type component interface {
	SetTablePrefix(prefix string)
	Get(name string) (bool, error)
}

func pageRowHandler() func(w http.ResponseWriter, r *http.Request) {
	return interfaceRowHandler(func() component {
		return &model.Page{}
	})
}

func menuRowHandler() func(w http.ResponseWriter, r *http.Request) {
	return interfaceRowHandler(func() component {
		return &model.Menu{}
	})
}

func blockInterfaceRowHandler() func(w http.ResponseWriter, r *http.Request) {
	return interfaceRowHandler(func() component {
		return &model.BlockInterface{}
	})
}

func interfaceRowHandler(fn func() component) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		logger := getLogger(r)
		client := getClient(r)

		c := fn()
		c.SetTablePrefix(client.Prefix())
		if ok, err := c.Get(params[keyName]); err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting one row")
			errorResponse(w, errQuery, http.StatusInternalServerError)
			return
		} else if !ok {
			errorResponse(w, errNotFound, http.StatusInternalServerError)
			return
		}

		jsonResponse(w, c)
	}
}
