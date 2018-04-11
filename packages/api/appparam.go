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
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

func appParam(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	_, prefix, err := checkEcosystem(w, data, logger)
	if err != nil {
		return err
	}
	ap := &model.AppParam{}
	ap.SetTablePrefix(prefix)
	found, err := ap.Get(nil, converter.StrToInt64(data.params[`appid`].(string)), data.params[`name`].(string))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting app parameter by name")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": data.params["name"].(string)}).Error("app parameter not found")
		return errorAPI(w, err, http.StatusBadRequest)
	}

	data.result = &paramValue{ID: converter.Int64ToStr(ap.ID), Name: ap.Name, Value: ap.Value,
		Conditions: ap.Conditions}
	return
}
