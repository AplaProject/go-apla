// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/gorilla/schema"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	KeyID         int64
	EcosystemID   int64
	EcosystemName string
	RoleID        int64
	IsMobile      string
	IsVDE         bool
}

// Prefix returns prefix of ecosystem
func (c *Client) Prefix() (prefix string) {
	prefix = converter.Int64ToStr(c.EcosystemID)
	if c.IsVDE {
		prefix += `_vde`
	}
	return
}

type form struct{}

func (f *form) Validate(r *http.Request) error {
	return nil
}

type formValidater interface {
	Validate(r *http.Request) error
}

type hexValue struct {
	value []byte
}

func (hv hexValue) Value() []byte {
	return hv.value
}

func (hv *hexValue) UnmarshalText(v []byte) (err error) {
	hv.value, err = hex.DecodeString(string(v))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "value": string(v), "error": err}).Error("decoding from hex")
	}
	return
}

type ecosystemForm struct {
	EcosystemID     int64  `schema:"ecosystem"`
	EcosystemPrefix string `schema:"-"`
}

func (f *ecosystemForm) Validate(r *http.Request) error {
	return f.ValidateEcosystem(r)
}

func (f *ecosystemForm) ValidateEcosystem(r *http.Request) error {
	client := getClient(r)
	logger := getLogger(r)

	if f.EcosystemID > 0 {
		count, err := model.GetNextID(nil, "1_ecosystems")
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id of ecosystems")
			return newError(err, http.StatusInternalServerError)
		}
		if f.EcosystemID >= count {
			logger.WithFields(log.Fields{"state_id": f.EcosystemID, "count": count, "type": consts.ParameterExceeded}).Error("ecosystem is larger then max count")
			return errEcosystem.Errorf(f.EcosystemID)
		}
	} else {
		f.EcosystemID = client.EcosystemID
	}

	f.EcosystemPrefix = converter.Int64ToStr(f.EcosystemID)
	if client.IsVDE {
		f.EcosystemPrefix += `_vde`
	}

	return nil
}

func parseForm(r *http.Request, f formValidater) error {
	r.ParseForm()
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(f, r.Form); err != nil {
		return newError(err, http.StatusBadRequest)
	}
	return f.Validate(r)
}

func jsonResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}
