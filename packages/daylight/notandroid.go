// +build !android,!ios

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

package daylight

import (
	"net"
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"

	log "github.com/sirupsen/logrus"
)

func httpListener(ListenHTTPHost string, route http.Handler) {
	l, err := net.Listen("tcp", ListenHTTPHost)
	log.WithFields(log.Fields{"host": ListenHTTPHost, "type": consts.NetworkError}).Debug("trying to listen at")
	if err == nil {
		log.WithFields(log.Fields{"host": ListenHTTPHost}).Info("listening at")
	} else {
		log.WithFields(log.Fields{"host": ListenHTTPHost, "error": err, "type": consts.NetworkError}).Debug("cannot listen at host")
	}

	go func() {
		srv := &http.Server{Handler: route}
		err = srv.Serve(l)
		if err != nil {
			log.WithFields(log.Fields{"host": ListenHTTPHost, "error": err, "type": consts.NetworkError}).Fatal("serving http at host")
			panic(err)
		}
	}()
}
