package api

import (
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/modes"
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

	ecosysID, err := modes.ValidateEcosysID(f.EcosystemID, client.EcosystemID, logger)
	if err != nil {
		if err == modes.ErrEcosystemNotFound {
			err = errEcosystem.Errorf(f.EcosystemID)
		}
		return err
	}

	f.EcosystemID = ecosysID
	f.EcosystemPrefix = converter.Int64ToStr(f.EcosystemID)

	return nil
}
