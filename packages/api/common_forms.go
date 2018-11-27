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

	ecosysID, err := validateEcosysID(f.EcosystemID, client.EcosystemID, logger)
	if err != nil {
		return err
	}

	f.EcosystemID = ecosysID
	f.EcosystemPrefix = converter.Int64ToStr(f.EcosystemID)

	return nil
}

func validateEcosysID(formID, clientID int64, logger *log.Entry) (int64, error) {
	if conf.Config.IsSupportingOBS() {
		return consts.DefaultOBS, nil
	}

	if formID <= 0 {
		return clientID, nil
	}

	count, err := model.GetNextID(nil, "1_ecosystems")
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id of ecosystems")
		return 0, err
	}

	if formID >= count {
		logger.WithFields(log.Fields{"state_id": formID, "count": count, "type": consts.ParameterExceeded}).Error("ecosystem is larger then max count")
		return 0, errEcosystem.Errorf(formID)
	}

	return formID, nil
}
