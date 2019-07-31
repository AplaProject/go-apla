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
	"strings"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/publisher"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type configOptionHandler func(w http.ResponseWriter, option string) error

func getConfigOptionHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	if len(params["option"]) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject, "error": "option not specified"}).Error("on getting option in config handler")
		errorResponse(w, errNotFound)
		return
	}

	switch params["option"] {
	case "centrifugo":
		centrifugoAddressHandler(w, r)
		return
	}

	errorResponse(w, errNotFound)
}

func replaceHttpSchemeToWs(centrifugoURL string) string {
	if strings.HasPrefix(centrifugoURL, "http:") {
		return strings.Replace(centrifugoURL, "http:", "ws:", -1)
	} else if strings.HasPrefix(centrifugoURL, "https:") {
		return strings.Replace(centrifugoURL, "https:", "wss:", -1)
	}
	return centrifugoURL
}

func centrifugoAddressHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	if _, err := publisher.GetStats(); err != nil {
		logger.WithFields(log.Fields{"type": consts.CentrifugoError, "error": err}).Warn("on getting centrifugo stats")
		errorResponse(w, err)
		return
	}

	jsonResponse(w, replaceHttpSchemeToWs(conf.Config.Centrifugo.URL))
}
