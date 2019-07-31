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
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type componentModel interface {
	SetTablePrefix(prefix string)
	Get(name string) (bool, error)
}

func getPageRowHandler(w http.ResponseWriter, r *http.Request) {
	getInterfaceRow(w, r, &model.Page{})
}

func getMenuRowHandler(w http.ResponseWriter, r *http.Request) {
	getInterfaceRow(w, r, &model.Menu{})
}

func getBlockInterfaceRowHandler(w http.ResponseWriter, r *http.Request) {
	getInterfaceRow(w, r, &model.BlockInterface{})
}

func getInterfaceRow(w http.ResponseWriter, r *http.Request, c componentModel) {
	params := mux.Vars(r)
	logger := getLogger(r)
	client := getClient(r)

	c.SetTablePrefix(client.Prefix())
	if ok, err := c.Get(params["name"]); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting one row")
		errorResponse(w, errQuery)
		return
	} else if !ok {
		errorResponse(w, errNotFound)
		return
	}

	jsonResponse(w, c)
}
