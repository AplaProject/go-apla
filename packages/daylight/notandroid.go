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
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/tcpserver"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	log "github.com/sirupsen/logrus"
)

func httpListener(ListenHTTPHost string, BrowserHTTPHost *string, route http.Handler) error {
	i := 0
	host := ListenHTTPHost
	var l net.Listener
	var err error
	for {
		i++
		if i > 7 {
			log.Warning("tried to listen ipV4 at all ports")
			return fmt.Errorf("tried all ports")
		}
		if i > 1 {
			host = ":7" + converter.IntToStr(i) + "79"
			*BrowserHTTPHost = "http://" + host
		}
		l, err = net.Listen("tcp4", host)
		log.WithFields(log.Fields{"host": host}).Debug("trying to listen at")
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
			log.WithFields(log.Fields{"host": host, "error": err}).Fatal("serving http at host")
		}
	}()
	return nil
}

// For ipv6 on the server
func httpListenerV6(route http.Handler) error {
	i := 0
	port := *utils.ListenHTTPPort
	var l net.Listener
	var err error
	for {
		if i > 7 {
			log.Error("tried all ports")
			return fmt.Errorf("tried all ports")
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
			log.WithFields(log.Fields{"error": err, "host": ":" + port}).Fatal("serving http at host")
		}
	}()
	return nil
}

func tcpListener() {
	go func() {
		log.WithFields(log.Fields{"host": *utils.TCPHost}).Info("Starting tcp listener at host")
		l, err := net.Listen("tcp4", *utils.TCPHost+":"+consts.TCP_PORT)
		if err != nil {
			log.WithFields(log.Fields{"host": *utils.TCPHost, "error": err, "type": consts.NetworkError}).Error("Error tcp listening at host")
		} else {
			go func() {
				for {
					conn, err := l.Accept()
					if err != nil {
						log.WithFields(log.Fields{"host": *utils.TCPHost, "error": err, "type": consts.NetworkError}).Error("Error accepting tcp at host")
						time.Sleep(time.Second)
					} else {
						go func(conn net.Conn) {
							tcpserver.HandleTCPRequest(conn)
							conn.Close()
						}(conn)
					}
				}
			}()
		}
	}()
}
