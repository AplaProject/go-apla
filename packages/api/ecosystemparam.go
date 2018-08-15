// MIT License
//
// Copyright (c) 2016 GenesisCommunity
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package api

import (
	"net/http"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

func ecosystemParam(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	_, prefix, err := checkEcosystem(w, data, logger)
	if err != nil {
		return err
	}
	sp := &model.StateParameter{}
	sp.SetTablePrefix(prefix)
	name := data.params[`name`].(string)
	found, err := sp.Get(nil, name)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting state parameter by name")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": name}).Error("state parameter not found")
		return errorAPI(w, `E_PARAMNOTFOUND`, http.StatusBadRequest, name)
	}

	data.result = &paramValue{ID: converter.Int64ToStr(sp.ID), Name: sp.Name, Value: sp.Value, Conditions: sp.Conditions}
	return
}

func getEcosystemName(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	ecosystemID := data.params["id"].(int64)
	ecosystems := model.Ecosystem{}
	found, err := ecosystems.Get(ecosystemID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting ecosystem name")
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "ecosystem_id": ecosystemID}).Error("ecosystem by id not found")
		return errorAPI(w, `E_PARAMNOTFOUND`, http.StatusNotFound, "name")
	}

	data.result = &struct {
		EcosystemName string `json:"ecosystem_name"`
	}{
		EcosystemName: ecosystems.Name,
	}
	return nil
}
