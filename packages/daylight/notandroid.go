// +build !android,!ios

// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
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

package daylight

import (
	"net"
	"net/http"
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"

	log "github.com/sirupsen/logrus"
)

func httpListener(ListenHTTPHost string, route http.Handler) {
	l, err := net.Listen("tcp4", ListenHTTPHost)
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

// For ipv6 on the server
func httpListenerV6(route http.Handler) {
	i := 0
	port := strconv.Itoa(conf.Config.HTTP.Port)
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
