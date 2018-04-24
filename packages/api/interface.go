package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	log "github.com/sirupsen/logrus"
)

type componentModel interface {
	SetTablePrefix(prefix string)
	Get(name string) (bool, error)
}

func getPageRow(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	return getInterfaceRow(w, r, data, logger, &model.Page{})
}

func getMenuRow(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	return getInterfaceRow(w, r, data, logger, &model.Menu{})
}

func getBlockInterfaceRow(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	return getInterfaceRow(w, r, data, logger, &model.BlockInterface{})
}

func getInterfaceRow(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry, c componentModel) error {
	c.SetTablePrefix(getPrefix(data))
	ok, err := c.Get(data.ParamString("name"))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting one row")
		return errorAPI(w, `E_QUERY`, http.StatusInternalServerError)
	}
	if !ok {
		return errorAPI(w, `E_NOTFOUND`, http.StatusInternalServerError)
	}

	data.result = c

	return nil
}
