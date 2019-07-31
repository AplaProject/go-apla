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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type contractsResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

func getContractsHandler(w http.ResponseWriter, r *http.Request) {
	form := &paginatorForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	client := getClient(r)
	logger := getLogger(r)

	contract := &model.Contract{}
	contract.EcosystemID = client.EcosystemID

	count, err := contract.CountByEcosystem()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting table records count")
		errorResponse(w, err)
		return
	}

	contracts, err := contract.GetListByEcosystem(form.Offset, form.Limit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all")
		errorResponse(w, err)
		return
	}

	list := make([]map[string]string, len(contracts))
	for i, c := range contracts {
		list[i] = c.ToMap()
		list[i]["address"] = converter.AddressToString(c.WalletID)
	}

	if len(list) == 0 {
		list = nil
	}

	jsonResponse(w, &listResult{
		Count: converter.Int64ToStr(count),
		List:  list,
	})
}
