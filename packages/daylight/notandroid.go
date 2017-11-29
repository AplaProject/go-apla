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

	conf "github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"

	log "github.com/sirupsen/logrus"
)

func httpListener(ListenHTTPHost string, BrowserHTTPHost *string, route http.Handler) {
	i := 0
	host := ListenHTTPHost
	var l net.Listener
	var err error
	for {
		i++
		if i > 7 {
			log.Warning("tried to listen ipV4 at all ports")
			panic("Error listening ")
		}
		if i > 1 {
			host = ":7" + converter.IntToStr(i) + "79"
			*BrowserHTTPHost = "http://" + host
		}
		l, err = net.Listen("tcp4", host)
		log.WithFields(log.Fields{"host": host, "type": consts.NetworkError}).Debug("trying to listen at")
		if err == nil {
			log.WithFields(log.Fields{"host": host}).Info("listening at")
			break
		} else {
			log.WithFields(log.Fields{"host": host, "error": err, "type": consts.NetworkError}).Debug("cannot listen at host")
		}
	}

	go func() {
		srv := &http.Server{Handler: route}
		err = srv.Serve(l)
		if err != nil {
			log.WithFields(log.Fields{"host": host, "error": err, "type": consts.NetworkError}).Fatal("serving http at host")
			panic(err)
		}
	}()
}

// For ipv6 on the server
func httpListenerV6(route http.Handler) {
	i := 0
	port := conf.Config.HTTPPort
	var l net.Listener
	var err error
	for {
		if i > 7 {
			log.WithFields(log.Fields{"type": consts.NetworkError}).Error("tried all ports")
			panic("Error listening ")
		}
		if i > 0 {
			port = "7" + converter.IntToStr(i) + "79"
		}
		i++
		l, err = net.Listen("tcp6", ":"+port)
		if err == nil {
			log.WithFields(log.Fields{"host": ":" + port}).Info("listening ipv6 at")
			break
		} else {
			log.WithFields(log.Fields{"error": err, "host": ":" + port, "type": consts.NetworkError}).Error("cannot listenin at host")
		}
	}

	go func() {
		srv := &http.Server{Handler: route}
		err = srv.Serve(l)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "host": ":" + port}).Error("serving http at host")
			panic(err)
		}
	}()
}
