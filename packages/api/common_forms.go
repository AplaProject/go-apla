package api

import (
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	log "github.com/sirupsen/logrus"
)

const (
	defaultPaginatorLimit = 25
	maxPaginatorLimit     = 1000
)

type paginatorForm struct {
	defaultLimit int64

	Limit  int64 `schema:"limit"`
	Offset int64 `schema:"offset"`
}

func (f *paginatorForm) Validate(r *http.Request) error {
	if f.Limit <= 0 {
		f.Limit = f.defaultLimit
		if f.Limit == 0 {
			f.Limit = defaultPaginatorLimit
		}
	}

	if f.Limit > maxPaginatorLimit {
		f.Limit = maxPaginatorLimit
	}

	return nil
}

type paramsForm struct {
	nopeValidator
	Names string `schema:"names"`
}

func (f *paramsForm) AcceptNames() map[string]bool {
	names := make(map[string]bool)
	for _, item := range strings.Split(f.Names, ",") {
		if len(item) == 0 {
			continue
		}
		names[item] = true
	}
	return names
}

type ecosystemForm struct {
	EcosystemID     int64  `schema:"ecosystem"`
	EcosystemPrefix string `schema:"-"`
}

func (f *ecosystemForm) Validate(r *http.Request) error {
	client := getClient(r)
	logger := getLogger(r)

	if conf.Config.IsSupportingOBS() {
		f.EcosystemID = consts.DefaultOBS
	} else if f.EcosystemID > 0 {
		count, err := model.GetNextID(nil, "1_ecosystems")
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id of ecosystems")
			return err
		}
		if f.EcosystemID >= count {
			logger.WithFields(log.Fields{"state_id": f.EcosystemID, "count": count, "type": consts.ParameterExceeded}).Error("ecosystem is larger then max count")
			return errEcosystem.Errorf(f.EcosystemID)
		}
	} else {
		f.EcosystemID = client.EcosystemID
	}

	f.EcosystemPrefix = converter.Int64ToStr(f.EcosystemID)

	return nil
}
