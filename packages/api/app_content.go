// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package api

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type appContentResult struct {
	Blocks    []model.BlockInterface `json:"blocks"`
	Pages     []model.Page           `json:"pages"`
	Contracts []model.Contract       `json:"contracts"`
}

func (m Mode) getAppContentHandler(w http.ResponseWriter, r *http.Request) {
	form := &appParamsForm{
		ecosystemForm: ecosystemForm{
			Validator: m.EcosysIDValidator,
		},
	}

	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	logger := getLogger(r)
	params := mux.Vars(r)

	bi := &model.BlockInterface{}
	p := &model.Page{}
	c := &model.Contract{}
	appID := converter.StrToInt64(params["appID"])
	ecosystemID := converter.StrToInt64(form.EcosystemPrefix)

	blocks, err := bi.GetByApp(appID, ecosystemID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting block interfaces by appID")
		errorResponse(w, err)
		return
	}

	pages, err := p.GetByApp(appID, ecosystemID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting pages by appID")
		errorResponse(w, err)
		return
	}

	contracts, err := c.GetByApp(appID, ecosystemID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting pages by appID")
		errorResponse(w, err)
		return
	}

	jsonResponse(w, &appContentResult{
		Blocks:    blocks,
		Pages:     pages,
		Contracts: contracts,
	})
}
